package cache

import (
	"testing"

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

func TestCacheSetGet(t *testing.T) {
	c := NewCache(nil)
	res := msgTestMiekMx()
	mt, opt := classify(res)
	do := false
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

	name := middleware.Name(res.Question[0].Name).Normalize()
	nxdomain := nameErrorKey(name, do)
	if i, ok := c.cache.Get(nxdomain); ok {
		resp := i.(*item).toMsg(res)
		t.Logf("%s\n", resp.String())
	}

	qtype := res.Question[0].Qtype
	successOrNoData := successKey(name, qtype, do)
	if i, ok := c.cache.Get(successOrNoData); ok {
		resp := i.(*item).toMsg(res)
		t.Logf("%s\n", resp.String())
	}
}

// TODO
func TestClassify(t *testing.T) {

}
