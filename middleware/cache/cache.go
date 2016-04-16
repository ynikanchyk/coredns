package cache

/*
The idea behind this implementation is as follows.  We keep a 2 small cache
withs Items that are indexed:

- negative cache: bit with qname only for NXDOMAIN responses
- negative cache: bit qname + qtype for NODATA responses
- positive cache: bit qname + qtype for succesful responses.

We track DNSSEC responses separately, i.e. under a different cache key.
Each Item stored contains the message split up in the different sections
and a few bits of the msg header.

For instance an NXDOMAIN for blaat.miek.nl will create the
following negative cache entry (do signal state of DO (do off, DO on)).

	ncache: do <blaat.miek.nl>
	Item:
		Ns: <miek.nl> SOA RR

If found a return packet is assembled and returned to the client. Taking size and EDNS0
constraints into account.

We also need to track if the answer received was an authoritative answer, ad bit and other
setting, for this we also store the header bits.

For the positive cache we use the same idea. Truncated responses are never stored.
*/

import (
	"log"
	"strings"
	"time"

	"github.com/miekg/coredns/middleware"

	"github.com/miekg/dns"
	gcache "github.com/patrickmn/go-cache"
)

// Cache is middleware that looks up responses in a cache and caches replies.
type Cache struct {
	Next  middleware.Handler
	cache *gcache.Cache
}

func NewCache(next middleware.Handler) Cache {
	return Cache{Next: next, cache: gcache.New(defaultDuration, purgeDuration)}
}

type MessageType int

const (
	Success    MessageType = iota
	NameError              // NXDOMAIN in header, SOA in auth.
	NoData                 // NOERROR in header, SOA in auth.
	OtherError             // Don't cache these
)

// classify classifies a message, it returns the MessageType.
func classify(m *dns.Msg) (MessageType, *dns.OPT) {
	opt := m.IsEdns0()
	soa := false
	if m.Rcode == dns.RcodeSuccess {
		return Success, opt
	}
	for _, r := range m.Ns {
		if r.Header().Rrtype == dns.TypeSOA {
			soa = true
			break
		}
	}

	// Check length of different section, and drop stuff that is just to large.
	if soa && m.Rcode == dns.RcodeSuccess {
		return NoData, opt
	}
	if soa && m.Rcode == dns.RcodeNameError {
		return NameError, opt
	}

	return OtherError, opt
}

func cacheKey(m *dns.Msg, t MessageType, do bool) string {
	if m.Truncated {
		return ""
	}

	qtype := m.Question[0].Qtype
	qname := strings.ToLower(m.Question[0].Name)
	switch t {
	case Success:
		return successKey(qname, qtype, do)
	case NameError:
		return nameErrorKey(qname, do)
	case NoData:
		return noDataKey(qname, qtype, do)
	case OtherError:
		return ""
	}
	return ""
}

type CachingResponseWriter struct {
	dns.ResponseWriter
	cache *gcache.Cache
}

func NewCachingResponseWriter(w dns.ResponseWriter, cache *gcache.Cache) *CachingResponseWriter {
	return &CachingResponseWriter{w, cache}
}

func (c *CachingResponseWriter) WriteMsg(res *dns.Msg) error {
	do := false
	mt, opt := classify(res)
	if opt != nil {
		do = opt.Do()
	}

	key := cacheKey(res, mt, do)
	if key != "" {
		i := newItem(res)
		switch mt {
		case Success:
			duration := MinTtl(res.Answer, mt)
			c.cache.Set(key, i, duration)
		case NameError, NoData:
			duration := MinTtl(res.Ns, mt)
			c.cache.Set(key, i, duration)
		}
	}

	return c.ResponseWriter.WriteMsg(res)
}

func (c *CachingResponseWriter) Write(buf []byte) (int, error) {
	log.Printf("[WARNING] Caching called with Write: not caching reply")
	n, err := c.ResponseWriter.Write(buf)
	return n, err
}

func (c *CachingResponseWriter) Hijack() {
	c.ResponseWriter.Hijack()
	return
}

func MinTtl(rrs []dns.RR, mt MessageType) time.Duration {
	if mt != Success || mt != NameError || mt != NoData {
		return 0
	}

	minTtl := maxTtl
	for _, r := range rrs {
		switch mt {
		case Success:
			if r.Header().Rrtype == dns.TypeSOA {
				return time.Duration(r.(*dns.SOA).Minttl) * time.Second
			}
		case NameError, NoData:
			if r.Header().Ttl < minTtl {
				minTtl = r.Header().Ttl
			}
		}
	}
	return time.Duration(minTtl) * time.Second
}

const (
	purgeDuration          = 1 * time.Minute
	defaultDuration        = 20 * time.Minute
	baseTtl                = 5 // minimum ttl that we will allow
	maxTtl          uint32 = 2 * 3600
)
