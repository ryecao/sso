package groups

import (
	"sync/atomic"
	"time"

	"golang.org/x/sync/syncmap"
)

// NewLocalCache returns a LocalCache instance
func NewLocalCache(
	ttl time.Duration,
	////logger logrus.FieldLogger,
	////inst metrics.Writer,
) *LocalCache {
	return &LocalCache{
		ttl:            ttl,
		localCacheData: &syncmap.Map{},
		////logger:     logger,
		////inst:       inst,
	}
}

type LocalCache struct {
	// Configuration
	length         uint64
	ttl            time.Duration
	localCacheData *syncmap.Map
	////logger logrus.FieldLogger
	////inst   metrics.Writer
	Entry
}

type Entry struct {
	Key  string
	Data []string
}

// retrieves an entry from the cache
func (lc *LocalCache) get(key string) ([]string, bool) {
	data, found := lc.localCacheData.Load(key)
	if data != nil {
		return data.([]string), found
	}
	return nil, false
}

// set will attempt to set some data to a given key, for the prescribed
// TTL and only if there is space available (it will not evict to make room)
func (lc *LocalCache) set(key string, data []string) error {
	if len(data) == 0 {
		return nil
	}

	// set the key
	lc.localCacheData.Store(key, data)
	atomic.AddUint64(&lc.length, 1)

	// Spawn the TTL cleanup goroutine if a TTL is set
	if lc.ttl > 0 {
		go func(key string) {
			<-time.After(lc.ttl)
			lc.Purge([]string{key})
		}(key)
	}
	return nil
}

// Retrieves a key from a local cache. If the key is not found, it will
// try to grab it from upstream and also attempt to cache it locally if
// that was successful
func (lc *LocalCache) Get(keys []string) ([]Entry, error) {
	entries := make([]Entry, 0, len(keys))
	for _, key := range keys {
		data, found := lc.get(key)
		if found {
			entries = append(entries, Entry{
				Key:  key,
				Data: data,
			})
		}
	}
	return entries, nil
}

// Set will set a number of entries within the current cache
func (lc *LocalCache) Set(entries []Entry) ([]Entry, error) {
	for _, entry := range entries {
		if err := lc.set(entry.Key, entry.Data); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

// Purge will remove a set of keys from the local cache map
func (lc *LocalCache) Purge(keys []string) error {
	for _, key := range keys {
		lc.localCacheData.Delete(key)
		atomic.AddUint64(&lc.length, ^uint64(0))
	}
	return nil
}

func (*LocalCache) String() string {
	return "LocalCache"
}
