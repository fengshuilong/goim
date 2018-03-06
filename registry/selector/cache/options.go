package cache

import (
	"time"

	"registry/selector"
	"golang.org/x/net/context"
)

type ttlKey struct{}

// Set the cache ttl
func TTL(t time.Duration) selector.Option {
	return func(o *selector.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, ttlKey{}, t)
	}
}
