package cwlogs

import (
	"sync"
	"time"
)

const (
	defaultMaxSize = 50000
	defaultTTL     = 60 * time.Second
)

type AfterAddFunc func()

type Deduplicator interface {
	GetLastTimestamp() int64
	AddAndExecuteIfNotPresent(eventID string, timestamp int64, afterAdd AfterAddFunc)
}

type deduplicatorImpl struct {
	timeToLive    time.Duration
	sizeLimit     int
	ids           map[string]int64
	lastTimestamp int64
	sync.RWMutex
}

func (d *deduplicatorImpl) evictOld() {
	oldestAlive := time.Now().Add(-d.timeToLive).Unix()

	for k, v := range d.ids {
		if v <= oldestAlive {
			delete(d.ids, k)
		}

	}
}

func (d *deduplicatorImpl) add(eventID string, timestamp int64) {
	if len(d.ids)+1 >= d.sizeLimit {
		d.evictOld()
	}
	d.ids[eventID] = timestamp
	if timestamp > d.lastTimestamp {
		d.lastTimestamp = timestamp
	}
}

func (d *deduplicatorImpl) GetLastTimestamp() int64 {
	d.Lock()
	defer d.Unlock()
	return d.lastTimestamp
}

func (d *deduplicatorImpl) AddAndExecuteIfNotPresent(eventID string, timestamp int64, after AfterAddFunc) {
	d.Lock()
	defer d.Unlock()
	if _, ok := d.ids[eventID]; !ok {
		d.add(eventID, timestamp)
		after()
	}
}

func NewDeduplicator(sizeLimit int, timeToLive time.Duration) Deduplicator {
	var szLimit int
	var ttl time.Duration
	if sizeLimit <= 0 {
		szLimit = defaultMaxSize
	} else {
		szLimit = sizeLimit
	}
	if timeToLive <= 0 {
		ttl = defaultTTL
	} else {
		ttl = timeToLive
	}

	return &deduplicatorImpl{
		sizeLimit:  szLimit,
		timeToLive: ttl,
		ids:        make(map[string]int64),
	}
}
