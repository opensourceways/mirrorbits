// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package scan

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/op/go-logging"
	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/core"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/network"
	"github.com/opensourceways/mirrorbits/utils"
)

var (
	// ErrScanAborted is returned when a scan is aborted by the user
	ErrScanAborted = errors.New("scan aborted")
	// ErrScanInProgress is returned when a scan is started while another is already in progress
	ErrScanInProgress = errors.New("scan already in progress")
	// ErrNoSyncMethod is returned when no sync protocol is available
	ErrNoSyncMethod = errors.New("no suitable URL for the scan")

	log = logging.MustGetLogger("main")
)

// Scanner is the interface that all scanners must implement
type Scanner interface {
	Scan(url, identifier string, conn redis.Conn, stop <-chan struct{}) (core.Precision, error)
}

type scan struct {
	redis *database.Redis
	cache *mirrors.Cache

	conn        redis.Conn
	mirrorid    int
	filesTmpKey string
	count       int64
}

type ScanResult struct {
	MirrorID     int
	MirrorName   string
	FilesIndexed int64
	KnownIndexed int64
	Removed      int64
	TZOffsetMs   int64
}

// IsScanning returns true is a scan is already in progress for the given mirror
func IsScanning(conn redis.Conn, id int) (bool, error) {
	return redis.Bool(conn.Do("EXISTS", fmt.Sprintf("SCANNING_%d", id)))
}

// Scan starts a scan of the given mirror
func Scan(typ core.ScannerType, r *database.Redis, c *mirrors.Cache, url string, id int, stop <-chan struct{}) (*ScanResult, error) {
	// Connect to the database
	conn := r.Get()
	defer conn.Close()

	s := &scan{
		redis:    r,
		mirrorid: id,
		conn:     conn,
		cache:    c,
	}

	scanner := &HttpScanner{
		scan: s,
	}

	// Get the mirror name
	name, err := redis.String(conn.Do("HGET", "MIRRORS", id))
	if err != nil {
		return nil, err
	}

	// Try to acquire a lock so we don't have a scanning race
	// from different nodes.
	// Also make the key expire automatically in case our process
	// gets killed.
	lock := network.NewClusterLock(s.redis, fmt.Sprintf("SCANNING_%d", id), name)

	done, err := lock.Get()
	if err != nil {
		return nil, err
	} else if done == nil {
		return nil, ErrScanInProgress
	}

	defer lock.Release()

	s.setLastSync(conn, id, typ, 0, false)

	defer func(err *error) {
		if err != nil && *err != nil {
			mirrors.PushLog(r, mirrors.NewLogError(id, *err))
		}
	}(&err)

	conn.Send("MULTI")

	filesKey := fmt.Sprintf("MIRRORFILES_%d", id)
	s.filesTmpKey = fmt.Sprintf("MIRRORFILESTMP_%d", id)

	// Remove any left over
	conn.Send("DEL", s.filesTmpKey)

	var precision core.Precision
	t1 := time.Now()
	precision, filePath, err := scanner.Scan(url, name, conn, stop)
	log.Infof("[%s] scan files cost time = %4.8f s \n", name, time.Since(t1).Seconds())
	if err != nil {
		// Discard MULTI
		s.ScannerDiscard()

		// Remove the temporary key
		conn.Do("DEL", s.filesTmpKey)

		log.Warningf("[%s] Removing %s from mirror", name, filePath)
		conn.Send("SREM", fmt.Sprintf("FILEMIRRORS_%s", filePath), id)

		log.Errorf("[%s] %s", name, err.Error())
		return nil, err
	}

	// Exec multi
	s.ScannerCommit()

	// Get the list of files no more present on this mirror
	var toremove []interface{}
	toremove, err = redis.Values(conn.Do("SDIFF", filesKey, s.filesTmpKey))
	if err != nil {
		return nil, err
	}

	// Remove this mirror from the given file SET
	if len(toremove) > 0 {
		conn.Send("MULTI")
		for _, e := range toremove {
			log.Debugf("[%s] Removing %s from mirror", name, e)
			conn.Send("SREM", fmt.Sprintf("FILEMIRRORS_%s", e), id)
			conn.Send("DEL", fmt.Sprintf("FILEINFO_%d_%s", id, e))
			// Publish update
			database.SendPublish(conn, database.MIRROR_FILE_UPDATE, fmt.Sprintf("%d %s", id, e))

		}
		_, err = conn.Do("EXEC")
		if err != nil {
			return nil, err
		}
	}

	// Finally rename the temporary sets containing the list
	// of files for this mirror to the production key
	if s.count > 0 {
		_, err = conn.Do("RENAME", s.filesTmpKey, filesKey)
		if err != nil {
			return nil, err
		}
	}

	sinterKey := fmt.Sprintf("HANDLEDFILES_%d", id)

	// Count the number of files known on the remote end
	common, _ := redis.Int64(conn.Do("SINTERSTORE", sinterKey, "FILES", filesKey))

	if err != nil {
		return nil, err
	}

	s.setLastSync(conn, id, typ, precision, true)

	var tzoffset int64
	tzoffset, err = s.adjustTZOffset(name, precision)
	if err != nil {
		log.Warningf("Unable to check timezone shifts: %s", err)
	}

	log.Infof("[%s] Indexed %d files (%d known), %d removed", name, s.count, common, len(toremove))
	res := &ScanResult{
		MirrorID:     id,
		MirrorName:   name,
		FilesIndexed: s.count,
		KnownIndexed: common,
		Removed:      int64(len(toremove)),
		TZOffsetMs:   tzoffset,
	}

	return res, nil
}

func (s *scan) ScannerAddFile(f filesystem.FileData) {
	s.count++

	// Add all the files to a temporary key
	s.conn.Send("SADD", s.filesTmpKey, f.Path)

	// Mark the file as being supported by this mirror
	rk := fmt.Sprintf("FILEMIRRORS_%s", f.Path)
	s.conn.Send("SADD", rk, s.mirrorid)

	// Save the size of the current file found on this mirror
	ik := fmt.Sprintf("FILEINFO_%d_%s", s.mirrorid, f.Path)
	s.conn.Send("HMSET", ik, "size", f.Size, "modTime", f.ModTime)

	// Publish update
	database.SendPublish(s.conn, database.MIRROR_FILE_UPDATE, fmt.Sprintf("%d %s", s.mirrorid, f.Path))
}

func (s *scan) ScannerDiscard() {
	s.conn.Do("DISCARD")
}

func (s *scan) ScannerCommit() error {
	_, err := s.conn.Do("EXEC")
	return err
}

func (s *scan) setLastSync(conn redis.Conn, id int, protocol core.ScannerType, precision core.Precision, successful bool) error {
	now := time.Now().UTC().Unix()

	conn.Send("MULTI")

	// Set the last sync time
	conn.Send("HSET", fmt.Sprintf("MIRROR_%d", id), "lastSync", now)

	// Set the last successful sync time
	if successful {
		if precision == 0 {
			precision = core.Precision(time.Second)
		}

		conn.Send("HMSET", fmt.Sprintf("MIRROR_%d", id),
			"lastSuccessfulSync", now,
			"lastSuccessfulSyncProtocol", protocol,
			"lastSuccessfulSyncPrecision", precision)
	}

	_, err := conn.Do("EXEC")

	// Publish an update on redis
	database.Publish(conn, database.MIRROR_UPDATE, strconv.Itoa(id))

	return err
}

func (s *scan) adjustTZOffset(name string, precision core.Precision) (ms int64, err error) {
	type pair struct {
		local  filesystem.FileInfo
		remote filesystem.FileInfo
	}

	var filepaths []string
	var pairs []pair
	var offsetmap map[int64]int
	var commonOffsetFound bool

	if s.cache == nil {
		log.Error("Skipping timezone check: missing cache in instance")
		return
	}

	if GetConfig().FixTimezoneOffsets == false {
		// We need to reset any previous value already
		// stored in the database.
		goto finish
	}

	// Get 100 random files from the mirror
	filepaths, err = redis.Strings(s.conn.Do("SRANDMEMBER", fmt.Sprintf("HANDLEDFILES_%d", s.mirrorid), 100))
	if err != nil {
		return
	}

	pairs = make([]pair, 0, 100)

	// Get the metadata of each file
	for _, path := range filepaths {
		p := pair{}

		p.local, err = s.cache.GetFileInfo(path)
		if err != nil {
			return
		}

		p.remote, err = s.cache.GetFileInfoMirror(s.mirrorid, path)
		if err != nil {
			return
		}

		if p.remote.ModTime.IsZero() {
			// Invalid mod time
			continue
		}

		if p.local.Size != p.remote.Size {
			// File differ: comparing the modfile will fail
			continue
		}

		// Add the file to valid pairs
		pairs = append(pairs, p)
	}

	if len(pairs) < 10 || len(pairs) < len(filepaths)/2 {
		// Less than half the files we got have a size
		// match, this is very suspicious. Skip the
		// check and reset the offset in the db.
		goto warn
	}

	// Compute the diff between local and remote for those files
	offsetmap = make(map[int64]int)
	for _, p := range pairs {
		// Convert to millisecond since unix timestamp truncating to the available precision
		local := p.local.ModTime.Truncate(precision.Duration()).UnixNano() / int64(time.Millisecond)
		remote := p.remote.ModTime.Truncate(precision.Duration()).UnixNano() / int64(time.Millisecond)

		diff := local - remote
		offsetmap[diff]++
	}

	for k, v := range offsetmap {
		// Find the common offset (if any) of at least 90% of our subset
		if v >= int(float64(len(pairs))/100*90) {
			ms = k
			commonOffsetFound = true
			break
		}
	}

warn:
	if !commonOffsetFound {
		log.Warningf("[%s] Unable to guess the timezone offset", name)
	}

finish:
	// Store the offset in the database
	key := fmt.Sprintf("MIRROR_%d", s.mirrorid)
	_, err = s.conn.Do("HMSET", key, "tzoffset", ms)
	if err != nil {
		return
	}

	// Publish update
	database.Publish(s.conn, database.MIRROR_UPDATE, strconv.Itoa(s.mirrorid))

	if ms != 0 {
		log.Noticef("[%s] Timezone offset detected: applied correction of %dms", name, ms)
	}

	return
}

type sourcescanner struct {
}

// Walk inside the source/reference repository
func (s *sourcescanner) walkSource(conn redis.Conn, d *filesystem.FileData) *filesystem.FileData {
	if d == nil {
		return nil
	}

	// Get the previous file properties
	properties, err := redis.Strings(conn.Do("HMGET", fmt.Sprintf("FILE_%s", d.Path), "size", "modTime", "sha256"))
	if err != nil && err != redis.ErrNil {
		log.Warningf("%s: get failed from redis: %s", d.Path, err.Error())
		return nil
	} else if len(properties) < 5 {
		// This will force a rehash
		properties = make([]string, 5)
	}

	size, _ := strconv.ParseInt(properties[0], 10, 64)
	modTime, _ := time.Parse(time.RFC1123, properties[1])
	sha256 := properties[2]

	if size != d.Size || !modTime.Equal(d.ModTime) || d.Sha256 != sha256 {
		log.Infof("[Old] %s: SIZE = %s, MODTIME = %s, SHA256 %s", d.Path, properties[0], modTime.String(), d.Sha256)
		log.Infof("[New] %s: SIZE = %s, MODTIME = %s, SHA256 %s", d.Path, strconv.FormatInt(d.Size, 10), d.ModTime.String(), d.Sha256)
	}
	return d
}

// ScanSource starts a scan of the local repository
func ScanSource(r *database.Redis, forceRehash bool, stop <-chan struct{}) (err error) {
	s := &sourcescanner{}

	conn := r.Get()
	defer conn.Close()

	if conn.Err() != nil {
		return conn.Err()
	}

	sourceFiles := make([]*filesystem.FileData, 0, 1024)

	//TODO lock atomically inside redis to avoid two simultaneous scan
	cnf := GetConfig()
	if _, err = os.Stat(cnf.Repository); os.IsNotExist(err) {
		return fmt.Errorf("%s: No such file or directory", cnf.Repository)
	}
	repoFileText := cnf.Repository + ".txt"
	if _, err = os.Stat(repoFileText); err != nil {
		return fmt.Errorf("%s: No such file or directory", repoFileText)
	}

	// open the file
	file, err := os.Open(repoFileText)
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// handle errors while opening
	if err != nil {
		return fmt.Errorf("cannot open the file: %s", repoFileText)
	}
	fileScanner := bufio.NewScanner(file)
	x, y, z := 0, 0, 0
	if fileScanner.Scan() {
		line := fileScanner.Bytes()
		z = len(line) // 47
		// drwxrwxrwx          4,096 2024/08/08 11:01:29 .
		//                           x                   y
		y = z - 1
		x = z - 21
	}
	log.Info("[source] Scanning the filesystem...")
	// read line by line
	for fileScanner.Scan() {
		line := fileScanner.Bytes()
		path := string(line[y:])
		if filesystem.Filter(path) && len(line) >= z {
			fd := filesystem.BuildFileTree(path, line[:x-1], line[x:y-1], cnf)
			fd = s.walkSource(conn, fd)
			if fd != nil {
				sourceFiles = append(sourceFiles, fd)
			}
		}
	}
	if err = fileScanner.Err(); err != nil {
		log.Errorf("Error while reading file: %s", err)
	}

	filesystem.UpdateFileTree(cnf.RepositoryFilter)

	if utils.IsStopped(stop) {
		return ErrScanAborted
	}
	log.Info("[source] Indexing the files...")

	lock := network.NewClusterLock(r, "SOURCE_REPO_SYNC", "source repository")

	retry := 10
	for {
		if retry == 0 {
			return ErrScanInProgress
		}
		done, err := lock.Get()
		if err != nil {
			return err
		} else if done != nil {
			break
		}
		time.Sleep(1 * time.Second)
		retry--
	}

	defer lock.Release()

	conn.Send("MULTI")

	// Remove any left over
	conn.Send("DEL", "FILES_TMP")

	// Add all the files to a temporary key
	count := 0
	for _, e := range sourceFiles {
		conn.Send("SADD", "FILES_TMP", e.Path)
		count++
	}

	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	// Do a diff between the sets to get the removed files
	toremove, err := redis.Values(conn.Do("SDIFF", "FILES", "FILES_TMP"))

	// Create/Update the files' hash keys with the fresh infos
	conn.Send("MULTI")
	for _, e := range sourceFiles {
		conn.Send("HMSET", fmt.Sprintf("FILE_%s", e.Path),
			"size", e.Size,
			"modTime", e.ModTime,
			"sha256", e.Sha256)

		// Publish update
		database.SendPublish(conn, database.FILE_UPDATE, e.Path)
	}

	// Remove old keys
	if len(toremove) > 0 {
		for _, e := range toremove {
			conn.Send("DEL", fmt.Sprintf("FILE_%s", e))

			// Publish update
			database.SendPublish(conn, database.FILE_UPDATE, fmt.Sprintf("%s", e))
		}
	}

	// Finally rename the temporary sets containing the list
	// of files to the production key
	conn.Send("RENAME", "FILES_TMP", "FILES")

	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	log.Infof("[source] Scanned %d files", count)

	return nil
}
