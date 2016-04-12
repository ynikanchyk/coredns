package etcd

import (
	"log"
	"math/rand"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/miekg/skydns/msg"
)

// Watch start the subzone watcher.
// TODO: startup function, we already know the zones and our etcd connection.
// TODO(miek): fix on restart
func (e Etcd) Watch() {
	if e.Stubmap == nil {
		return
	}
	if stub {
		s.UpdateStubZones()
		go func() {
			duration := 1 * time.Second
			var watcher etcd.Watcher

			// Just use the first configured
			// domain
			// const /dns/stub
			watcher = client.Watcher(msg.Path(config.Domain)+"/dns/stub/", &etcd.WatcherOptions{AfterIndex: 0, Recursive: true})

			for {
				_, err := watcher.Next(ctx)

				if err != nil {
					//
					log.Printf("skydns: stubzone update failed, sleeping %s + ~3s", duration)
					time.Sleep(duration + (time.Duration(rand.Float32() * 3e9))) // Add some random.
					duration *= 2
					if duration > 32*time.Second {
						duration = 32 * time.Second
					}
				} else {
					s.UpdateStubZones()
					log.Printf("skydns: stubzone update")
					duration = 1 * time.Second // reset
				}
			}
		}()
	}
}
