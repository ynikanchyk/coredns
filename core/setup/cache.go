package setup

import (
	"github.com/miekg/coredns/middleware"
	"github.com/miekg/coredns/middleware/cache"
)

// Cache sets up the root file path of the server.
func Cache(c *Controller) (middleware.Middleware, error) {
	return func(next middleware.Handler) middleware.Handler {
		return cache.NewCache(next)
	}, nil
}
