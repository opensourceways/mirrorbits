// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package mirrors

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/gomodule/redigo/redis"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/network"
	"github.com/opensourceways/mirrorbits/utils"
)

// Cache implements a local caching mechanism of type LRU for content available in the
// redis database that is automatically invalidated if the object is updated in Redis.
type Cache struct {
	r        *database.Redis
	fiCache  *LRUCache
	fmCache  *LRUCache
	mCache   *LRUCache
	fimCache *LRUCache

	mirrorUpdateEvent      chan string
	fileUpdateEvent        chan string
	mirrorFileUpdateEvent  chan string
	pubsubReconnectedEvent chan string
	invalidationEvent      chan string
}

type fileInfoValue struct {
	value filesystem.FileInfo
}

func (f *fileInfoValue) Size() int {
	return int(unsafe.Sizeof(f.value))
}

type fileMirrorValue struct {
	value []int
}

func (f *fileMirrorValue) Size() int {
	return cap(f.value)
}

type mirrorValue struct {
	value Mirror
}

func (f *mirrorValue) Size() int {
	return int(unsafe.Sizeof(f.value))
}

// NewCache constructs a new instance of Cache
func NewCache(r *database.Redis) *Cache {
	if r == nil || r.Pubsub == nil {
		return nil
	}

	c := &Cache{
		r: r,
	}

	// Create the LRU
	c.fiCache = NewLRUCache(1024000)
	c.fmCache = NewLRUCache(2048000)
	c.mCache = NewLRUCache(1024000)
	c.fimCache = NewLRUCache(4096000)

	// Create event channels
	c.mirrorUpdateEvent = make(chan string, 10)
	c.fileUpdateEvent = make(chan string, 10)
	c.mirrorFileUpdateEvent = make(chan string, 10)
	c.pubsubReconnectedEvent = make(chan string)

	c.invalidationEvent = make(chan string, 10)

	// Subscribe to events
	c.r.Pubsub.SubscribeEvent(database.MIRROR_UPDATE, c.mirrorUpdateEvent)
	c.r.Pubsub.SubscribeEvent(database.FILE_UPDATE, c.fileUpdateEvent)
	c.r.Pubsub.SubscribeEvent(database.MIRROR_FILE_UPDATE, c.mirrorFileUpdateEvent)
	c.r.Pubsub.SubscribeEvent(database.PUBSUB_RECONNECTED, c.pubsubReconnectedEvent)

	go func() {
		for {
			//FIXME add a close channel
			select {
			case data := <-c.mirrorUpdateEvent:
				c.mCache.Delete(data)
				select {
				case c.invalidationEvent <- data:
				default:
					// Non-blocking
				}
			case data := <-c.fileUpdateEvent:
				c.fiCache.Delete(data)
			case data := <-c.mirrorFileUpdateEvent:
				s := strings.SplitN(data, " ", 2)
				c.fmCache.Delete(s[1])
				c.fimCache.Delete(fmt.Sprintf("%s|%s", s[0], s[1]))
			case <-c.pubsubReconnectedEvent:
				c.Clear()
			}
		}
	}()

	return c
}

// Clear clears the local cache
func (c *Cache) Clear() {
	c.fiCache.Clear()
	c.fmCache.Clear()
	c.mCache.Clear()
	c.fimCache.Clear()
}

// GetMirrorInvalidationEvent returns a channel that contains ID of mirrors
// that have just been invalidated. This function is supposed to have only
// ONE reader and is made to avoid a race for MIRROR_UPDATE events between
// a mirror invalidation and a mirror being fetched from the cache.
func (c *Cache) GetMirrorInvalidationEvent() <-chan string {
	return c.invalidationEvent
}

// GetFileInfo returns file information for a given file either from the cache
// or directly from the database if the object is not yet stored in the cache.
func (c *Cache) GetFileInfo(path string) (f filesystem.FileInfo, err error) {
	v, ok := c.fiCache.Get(path)
	if ok {
		f = v.(*fileInfoValue).value
	} else {
		f, err = c.fetchFileInfo(path)
	}
	return
}

func (c *Cache) fetchFileInfo(path string) (f filesystem.FileInfo, err error) {
	rconn := c.r.Get()
	defer rconn.Close()
	f.Path = path // Path is not stored in the object instance in redis

	reply, err := redis.Strings(rconn.Do("HMGET", fmt.Sprintf("FILE_%s", path), "size", "modTime", "sha256"))
	if err != nil {
		return
	}

	f.Size, _ = strconv.ParseInt(reply[0], 10, 64)
	f.ModTime, _ = time.Parse(time.RFC1123, reply[1])
	f.Sha256 = reply[2]
	c.fiCache.Set(path, &fileInfoValue{value: f})
	return
}

// GetMirrors returns all the mirrors serving a given file either from the cache
// or directly from the database if the object is not yet stored in the cache.
func (c *Cache) GetMirrors(path string, clientInfo network.GeoIPRecord) (mirrors []Mirror, err error) {
	var mirrorsIDs []int
	v, ok := c.fmCache.Get(path)
	if ok {
		mirrorsIDs = v.(*fileMirrorValue).value
	}

	if len(mirrorsIDs) == 0 {
		mirrorsIDs, err = c.fetchFileMirrors(path)
		if err != nil {
			return
		}
	}
	mirrors = make([]Mirror, 0, len(mirrorsIDs))
	for _, id := range mirrorsIDs {
		var mirror Mirror
		var fileInfo filesystem.FileInfo
		v, ok := c.mCache.Get(strconv.Itoa(id))
		if ok {
			mirror = v.(*mirrorValue).value
		} else {
			//TODO execute missing items in a MULTI query
			mirror, err = c.fetchMirror(id)
			if err != nil {
				return
			}
		}
		v, ok = c.fimCache.Get(fmt.Sprintf("%d|%s", id, path))
		if ok {
			fileInfo = v.(*fileInfoValue).value
		} else {
			fileInfo, err = c.fetchFileInfoMirror(id, path)
			if err != nil {
				return
			}
		}
		if fileInfo.Size >= 0 {
			mirror.FileInfo = &fileInfo
		}

		// Add the path in the results so we can access it from the templates
		mirror.FileInfo.Path = path

		if clientInfo.IsValid() {
			mirror.Distance = utils.GetDistanceKm(clientInfo.Latitude,
				clientInfo.Longitude,
				mirror.Latitude,
				mirror.Longitude)
		} else {
			mirror.Distance = 0
		}
		mirrors = append(mirrors, mirror)
	}
	return
}

func (c *Cache) fetchFileMirrors(path string) (ids []int, err error) {
	rconn := c.r.Get()
	defer rconn.Close()
	ids, err = redis.Ints(rconn.Do("SMEMBERS", fmt.Sprintf("FILEMIRRORS_%s", path)))
	if err != nil {
		return
	}
	c.fmCache.Set(path, &fileMirrorValue{value: ids})
	return
}

func (c *Cache) fetchMirror(mirrorID int) (mirror Mirror, err error) {
	rconn := c.r.Get()
	defer rconn.Close()
	reply, err := redis.Values(rconn.Do("HGETALL", fmt.Sprintf("MIRROR_%d", mirrorID)))
	if err != nil {
		return
	}
	if len(reply) == 0 {
		err = redis.ErrNil
		return
	}
	err = redis.ScanStruct(reply, &mirror)
	if err != nil {
		return
	}
	mirror.Prepare()
	c.mCache.Set(strconv.Itoa(mirrorID), &mirrorValue{value: mirror})
	return
}

func (c *Cache) GetFileInfoMirror(mirrorID int, path string) (f filesystem.FileInfo, err error) {
	var fileInfo filesystem.FileInfo

	v, ok := c.fimCache.Get(fmt.Sprintf("%d|%s", mirrorID, path))
	if ok {
		fileInfo = v.(*fileInfoValue).value
	} else {
		fileInfo, err = c.fetchFileInfoMirror(mirrorID, path)
		if err != nil {
			return
		}
	}
	return fileInfo, nil
}

func (c *Cache) fetchFileInfoMirror(id int, path string) (f filesystem.FileInfo, err error) {
	rconn := c.r.Get()
	defer rconn.Close()
	f.Path = path // Path is not stored in the object instance in redis

	reply, err := redis.Strings(rconn.Do("HMGET", fmt.Sprintf("FILEINFO_%d_%s", id, path), "size", "modTime", "sha256"))
	if err != nil {
		return
	}

	// Note: as of today, only the size is stored by the scanners
	// all other fields are left blank.

	f.Size, _ = strconv.ParseInt(reply[0], 10, 64)
	f.ModTime, _ = time.Parse(time.DateTime, reply[1])
	f.Sha256 = reply[2]

	c.fimCache.Set(fmt.Sprintf("%d|%s", id, path), &fileInfoValue{value: f})
	return
}

// GetMirror returns all information about a given mirror either from the cache
// or directly from the database if the object is not yet stored in the cache.
func (c *Cache) GetMirror(id int) (mirror Mirror, err error) {
	v, ok := c.mCache.Get(strconv.Itoa(id))
	if ok {
		mirror = v.(*mirrorValue).value
	} else {
		mirror, err = c.fetchMirror(id)
		if err != nil {
			return
		}
	}
	return
}
