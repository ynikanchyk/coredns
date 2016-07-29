package setup

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/miekg/coredns/middleware"
	"github.com/miekg/coredns/middleware/kubernetes"
	"github.com/miekg/coredns/middleware/kubernetes/nametemplate"
)

const (
	defaultNameTemplate = "{service}.{namespace}.{zone}"
	defaultResyncPeriod = 5 * time.Minute
)

// Kubernetes sets up the kubernetes middleware.
func Kubernetes(c *Controller) (middleware.Middleware, error) {
	kubernetes, err := kubernetesParse(c)
	if err != nil {
		return nil, err
	}

	err = kubernetes.StartKubeCache()
	if err != nil {
		return nil, err
	}

	return func(next middleware.Handler) middleware.Handler {
		kubernetes.Next = next
		return kubernetes
	}, nil
}

func kubernetesParse(c *Controller) (kubernetes.Kubernetes, error) {
	var err error
	template := defaultNameTemplate

	k8s := kubernetes.Kubernetes{}

	k8s.ResyncPeriod = defaultResyncPeriod
	k8s.NameTemplate = new(nametemplate.NameTemplate)
	k8s.NameTemplate.SetTemplate(template)

	// TODO: expose resync period in Corefile

	for c.Next() {
		if c.Val() == "kubernetes" {
			zones := c.RemainingArgs()

			log.Printf("[debug] Zones: %v", zones)
			if len(zones) == 0 {
				k8s.Zones = c.ServerBlockHosts
				log.Printf("[debug] Zones(from ServerBlockHosts): %v", zones)
			} else {
				// Normalize requested zones
				k8s.Zones = kubernetes.NormalizeZoneList(zones)
			}

			middleware.Zones(k8s.Zones).FullyQualify()
			if k8s.Zones == nil || len(k8s.Zones) < 1 {
				err = errors.New("Zone name must be provided for kubernetes middleware.")
				log.Printf("[debug] %v\n", err)
				return kubernetes.Kubernetes{}, err
			}

			for c.NextBlock() {
				switch c.Val() {
				case "template":
					args := c.RemainingArgs()
					if len(args) != 0 {
						template := strings.Join(args, "")
						err = k8s.NameTemplate.SetTemplate(template)
						if err != nil {
							return kubernetes.Kubernetes{}, err
						}
					} else {
						log.Printf("[debug] 'template' keyword provided without any template value.")
						return kubernetes.Kubernetes{}, c.ArgErr()
					}
				case "namespaces":
					args := c.RemainingArgs()
					if len(args) != 0 {
						k8s.Namespaces = append(k8s.Namespaces, args...)
					} else {
						log.Printf("[debug] 'namespaces' keyword provided without any namespace values.")
						return kubernetes.Kubernetes{}, c.ArgErr()
					}
				}
			}
			return k8s, nil
		}
	}
	err = errors.New("Kubernetes setup called without keyword 'kubernetes' in Corefile")
	log.Printf("[ERROR] %v\n", err)
	return kubernetes.Kubernetes{}, err
}
