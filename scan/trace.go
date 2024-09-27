// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package scan

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/database"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/utils"
)

var (
	userAgentName = "User-Agent"
	userAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"

	// ErrNoTrace is returned when no trace file is found
	ErrNoTrace = errors.New("No trace file")
)

// Trace is the internal trace handler
type Trace struct {
	redis      *database.Redis
	httpClient *resty.Client
	stop       <-chan struct{}
}

// NewTraceHandler returns a new instance of the trace file handler.
// Trace files are used to compute the time offset between a mirror
// and the local repository.
func NewTraceHandler(redis *database.Redis, stop <-chan struct{}) *Trace {
	t := &Trace{
		redis: redis,
		stop:  stop,
	}

	t.httpClient = resty.New().RemoveProxy()

	return t
}

// GetLastUpdate connects in HTTP to the mirror to get the latest
// trace file and computes the offset of the mirror.
func (t *Trace) GetLastUpdate(mirror mirrors.Mirror) error {
	traceFile := GetConfig().TraceFileLocation

	if len(traceFile) == 0 {
		return ErrNoTrace
	}

	log.Debugf("Getting latest trace file for %s...", mirror.Name)

	resp, err := t.httpClient.R().SetHeader(userAgentName, userAgent).Get(utils.ConcatURL(mirror.HttpURL, traceFile))
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bufio.NewReader(resp.RawBody()))
	scanner.Split(bufio.ScanWords)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return err
	}

	timestamp, err := strconv.ParseInt(scanner.Text(), 10, 64)
	if err != nil {
		return err
	}

	conn := t.redis.Get()
	defer conn.Close()

	_, err = conn.Do("HSET", fmt.Sprintf("MIRROR_%d", mirror.ID), "lastModTime", timestamp)
	if err != nil {
		return err
	}

	// Publish an update on redis
	database.Publish(conn, database.MIRROR_UPDATE, strconv.Itoa(mirror.ID))

	log.Debugf("[%s] trace last sync: %s", mirror.Name, time.Unix(timestamp, 0))
	return nil
}
