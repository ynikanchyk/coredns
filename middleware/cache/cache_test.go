package cache

import (
	"testing"
	"time"

	"github.com/miekg/coredns/middleware"
	"github.com/miekg/coredns/middleware/test"

	"github.com/miekg/dns"
)

type cacheTestCase struct {
	test.Case
	AuthenticatedData  bool
	Authoritative      bool
	RecursionAvailable bool
	Truncated          bool
}

var cacheTestCases = []cacheTestCase{
	{
		RecursionAvailable: true, AuthenticatedData: true, Authoritative: true,
		Case: test.Case{
			Qname: "miek.nl.", Qtype: dns.TypeMX,
			Answer: []dns.RR{
				test.MX("miek.nl.	1800	IN	MX	1 aspmx.l.google.com."),
				test.MX("miek.nl.	1800	IN	MX	10 aspmx2.googlemail.com."),
				test.MX("miek.nl.	1800	IN	MX	10 aspmx3.googlemail.com."),
				test.MX("miek.nl.	1800	IN	MX	5 alt1.aspmx.l.google.com."),
				test.MX("miek.nl.	1800	IN	MX	5 alt2.aspmx.l.google.com."),
			},
		},
	},
}

func cacheMsg(m *dns.Msg, tc cacheTestCase) *dns.Msg {
	m.RecursionAvailable = tc.RecursionAvailable
	m.AuthenticatedData = tc.AuthenticatedData
	m.Authoritative = tc.Authoritative
	m.Truncated = tc.Truncated
	m.Answer = tc.Answer
	m.Ns = tc.Ns
	return m
}

func newTestCache() (Cache, *CachingResponseWriter) {
	c := NewCache(0, []string{"."}, nil)
	crr := NewCachingResponseWriter(nil, c.cache, time.Duration(0))
	return c, crr
}

func TestCacheSetGet(t *testing.T) {
	c, crr := newTestCache()

	for _, tc := range cacheTestCases {
		m := tc.Msg()
		m = cacheMsg(m, tc)

		mt, _ := classify(m)
		do := tc.Case.Do
		key := cacheKey(m, mt, do)
		crr.Set(m, key, mt)

		name := middleware.Name(m.Question[0].Name).Normalize()
		qtype := m.Question[0].Qtype

		if i, ok := c.Get(name, qtype, do); ok {
			resp := i.toMsg(m)
			t.Logf("%s\n", resp.String())
		}
	}
}
