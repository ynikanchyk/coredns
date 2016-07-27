package setup

import (
	"errors"
	"log"
	"strings"

	"github.com/miekg/coredns/middleware"
	"github.com/miekg/coredns/middleware/kubernetes"
	k8sc "github.com/miekg/coredns/middleware/kubernetes/k8sclient"
	"github.com/miekg/coredns/middleware/kubernetes/nametemplate"
	"github.com/miekg/coredns/middleware/proxy"
)

const (
	defaultK8sEndpoint  = "http://localhost:8080"
	defaultNameTemplate = "{service}.{namespace}.{zone}"
)

// Kubernetes sets up the kubernetes middleware.
func Kubernetes(c *Controller) (middleware.Middleware, error) {
	// TODO: Determine if subzone support required

	kubernetes, err := kubernetesParse(c)
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
	k8s := kubernetes.Kubernetes{
		Proxy: proxy.New([]string{}),
	}
	var (
		endpoints  = []string{defaultK8sEndpoint}
		template   = defaultNameTemplate
		namespaces = []string{}
	)

	k8s.APIConn = k8sc.NewK8sConnector(endpoints[0])
	k8s.NameTemplate = new(nametemplate.NameTemplate)
	k8s.NameTemplate.SetTemplate(template)

	// TODO: clean this parsing up
	for c.Next() {
		if c.Val() == "kubernetes" {
			zones := c.RemainingArgs()

			if len(zones) == 0 {
				k8s.Zones = c.ServerBlockHosts
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

			if c.NextBlock() {
				// TODO(miek): 2 switches?
				switch c.Val() {
				case "endpoint":
					args := c.RemainingArgs()
					if len(args) == 0 {
						return kubernetes.Kubernetes{}, c.ArgErr()
					}
					endpoints = args
					k8s.APIConn = k8sc.NewK8sConnector(endpoints[0])
				case "template":
					args := c.RemainingArgs()
					if len(args) == 0 {
						return kubernetes.Kubernetes{}, c.ArgErr()
					}
					template = strings.Join(args, "")
					err = k8s.NameTemplate.SetTemplate(template)
					if err != nil {
						return kubernetes.Kubernetes{}, err
					}
				case "namespaces":
					args := c.RemainingArgs()
					if len(args) == 0 {
						return kubernetes.Kubernetes{}, c.ArgErr()
					}
					namespaces = args
					k8s.Namespaces = append(k8s.Namespaces, namespaces...)
				}
				for c.Next() {
					switch c.Val() {
					case "template":
						args := c.RemainingArgs()
						if len(args) == 0 {
							return kubernetes.Kubernetes{}, c.ArgErr()
						}
						template = strings.Join(args, "")
						err = k8s.NameTemplate.SetTemplate(template)
						if err != nil {
							return kubernetes.Kubernetes{}, err
						}
					case "namespaces":
						args := c.RemainingArgs()
						if len(args) == 0 {
							return kubernetes.Kubernetes{}, c.ArgErr()
						}
						namespaces = args
						k8s.Namespaces = append(k8s.Namespaces, namespaces...)
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
