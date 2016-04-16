package cache

import (
	"testing"
	"time"

	"github.com/miekg/coredns/middleware"
	"github.com/miekg/coredns/middleware/test"

	"github.com/miekg/dns"
)

func msgTestMiekMx() *dns.Msg {
	m := new(dns.Msg)
	m.SetQuestion("miek.nl.", dns.TypeMX)
	m.RecursionAvailable = true
	m.AuthenticatedData = true
	m.Answer = []dns.RR{
		test.MX("miek.nl.	1800	IN	MX	1 aspmx.l.google.com."),
		test.MX("miek.nl.	1800	IN	MX	5 alt1.aspmx.l.google.com."),
		test.MX("miek.nl.	1800	IN	MX	5 alt2.aspmx.l.google.com."),
		test.MX("miek.nl.	1800	IN	MX	10 aspmx2.googlemail.com."),
		test.MX("miek.nl.	1800	IN	MX	10 aspmx3.googlemail.com."),
	}
	return m
}

func newTestCache() (Cache, *CachingResponseWriter) {
	c := NewCache(0, []string{"."}, nil)
	crr := NewCachingResponseWriter(nil, c.cache, time.Duration(0))
	return c, crr
}

func TestCacheSetGet(t *testing.T) {
	c, crr := newTestCache()
	res := msgTestMiekMx()
	mt, opt := classify(res)
	do := false
	if opt != nil {
		do = opt.Do()
	}

	key := cacheKey(res, mt, do)
	crr.Set(res, key, mt)

	// TODO(miek): make this somewhat better and loop through a few messages.
	name := middleware.Name(res.Question[0].Name).Normalize()
	qtype := res.Question[0].Qtype

	if i, ok := c.Get(name, qtype, do); ok {
		resp := i.toMsg(res)
		t.Logf("%s\n", resp.String())
	}

	time.Sleep(2 * time.Second)

	if i, ok := c.Get(name, qtype, do); ok {
		resp := i.toMsg(res)
		t.Logf("%s\n", resp.String())
	}
}

func TestCacheTruncated(t *testing.T) {
	res := msgTestMiekMx()
	res.Truncated = true
	mt, _ := classify(res)
	key := cacheKey(res, mt, true)
	if key != "" {
		t.Errorf("Truncated message should lead to empty cache key, got %s", key)
	}
}

// TODO(miek)
func TestClassify(t *testing.T) {

}
