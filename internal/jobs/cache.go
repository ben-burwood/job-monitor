package jobs

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// fetchTimeout bounds a single upstream fetch, which runs detached from the request that triggered it
const fetchTimeout = time.Minute

// ttlCache caches values per job ID for a fixed TTL. Concurrent fetches of the same ID are collapsed into one,
// run detached from the caller's context so a client disconnecting can't fail the fetch for everyone else.
type ttlCache[T any] struct {
	ttl     time.Duration
	entries sync.Map // map[string]ttlEntry[T]
	group   singleflight.Group
}

type ttlEntry[T any] struct {
	value     T
	expiresAt time.Time
}

// get returns the cached value for id, fetching it if missing, expired or force is set. Fetch errors are returned, not cached.
func (c *ttlCache[T]) get(ctx context.Context, id string, force bool, fetch func(ctx context.Context) (T, error)) (T, error) {
	if !force {
		if v, ok := c.entries.Load(id); ok {
			if entry := v.(ttlEntry[T]); time.Now().Before(entry.expiresAt) {
				return entry.value, nil
			}
		}
	}

	v, err, _ := c.group.Do(id, func() (any, error) {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), fetchTimeout)
		defer cancel()
		value, err := fetch(ctx)
		if err == nil {
			c.entries.Store(id, ttlEntry[T]{value: value, expiresAt: time.Now().Add(c.ttl)})
		}
		return value, err
	})
	value, _ := v.(T) // comma-ok: a nil interface T comes back as untyped nil
	return value, err
}
