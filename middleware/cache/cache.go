package cache

/*
The idea behind this implementation is as follows. We have a cache that is index
by a couple different keys, which allows use to have:

- negative cache: qname only for NXDOMAIN responses
- negative cache: qname + qtype for NODATA responses
- positive cache: qname + qtype for succesful responses.

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
setting, for this we also store a few header bits.

For the positive cache we use the same idea. Truncated responses are never stored.
*/

import (
	"log"
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

type messageType int

const (
	success    messageType = iota
	nameError              // NXDOMAIN in header, SOA in auth.
	noData                 // NOERROR in header, SOA in auth.
	otherError             // Don't cache these.
)

// classify classifies a message, it returns the MessageType.
func classify(m *dns.Msg) (messageType, *dns.OPT) {
	opt := m.IsEdns0()
	soa := false
	if m.Rcode == dns.RcodeSuccess {
		return success, opt
	}
	for _, r := range m.Ns {
		if r.Header().Rrtype == dns.TypeSOA {
			soa = true
			break
		}
	}

	// Check length of different section, and drop stuff that is just to large.
	if soa && m.Rcode == dns.RcodeSuccess {
		return noData, opt
	}
	if soa && m.Rcode == dns.RcodeNameError {
		return nameError, opt
	}

	return otherError, opt
}

func cacheKey(m *dns.Msg, t messageType, do bool) string {
	if m.Truncated {
		return ""
	}

	qtype := m.Question[0].Qtype
	qname := middleware.Name(m.Question[0].Name).Normalize()
	switch t {
	case success:
		return successKey(qname, qtype, do)
	case nameError:
		return nameErrorKey(qname, do)
	case noData:
		return noDataKey(qname, qtype, do)
	case otherError:
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
		switch mt {
		case success:
			duration := minTtl(res.Answer, mt)
			i := newItem(res, duration)
			c.cache.Set(key, i, duration)
		case nameError, noData:
			duration := minTtl(res.Ns, mt)
			i := newItem(res, duration)
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

func minTtl(rrs []dns.RR, mt messageType) time.Duration {
	if mt != success || mt != nameError || mt != noData {
		return 0
	}

	minTtl := maxTtl
	for _, r := range rrs {
		switch mt {
		case success:
			if r.Header().Rrtype == dns.TypeSOA {
				return time.Duration(r.(*dns.SOA).Minttl) * time.Second
			}
		case nameError, noData:
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
