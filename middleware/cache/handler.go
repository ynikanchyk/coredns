package cache

import (
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

// ServeDNS implements the middleware.Handler interface.
func (c Cache) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	do := false
	opt := r.IsEdns0()
	if opt != nil {
		do = opt.Do()
	}
	name := strings.ToLower(r.Question[0].Name)
	qtype := r.Question[0].Qtype

	nxdomain := nameErrorKey(name, do)
	if i, ok := c.cache.Get(nxdomain); ok {
		resp := i.(*Item).toMsg(r)
		w.WriteMsg(resp)
		return dns.RcodeSuccess, nil
	}

	successOrNoData := successKey(name, qtype, do)
	if i, ok := c.cache.Get(successOrNoData); ok {
		resp := i.(*Item).toMsg(r)
		w.WriteMsg(resp)
		return dns.RcodeSuccess, nil
	}

	crr := NewCachingResponseWriter(w, c.cache)
	return c.Next.ServeDNS(ctx, crr, r)
}
