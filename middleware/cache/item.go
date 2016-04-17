package cache

import (
	"strconv"

	"github.com/miekg/dns"
)

type item struct {
	Authoritative      bool
	AuthenticatedData  bool
	RecursionAvailable bool
	Answer             []dns.RR
	Ns                 []dns.RR
	Extra              []dns.RR
}

func newItem(m *dns.Msg) *item {
	i := new(item)
	i.Authoritative = m.Authoritative
	i.AuthenticatedData = m.AuthenticatedData
	i.RecursionAvailable = m.RecursionAvailable
	i.Answer = m.Answer
	i.Ns = m.Ns
	i.Extra = m.Extra

	return i
}

// toMsg turns i into a message, it tailers to reply to m.
func (i *item) toMsg(m *dns.Msg) *dns.Msg {
	m1 := new(dns.Msg)
	m1.SetReply(m)
	m1.Authoritative = i.Authoritative
	m1.AuthenticatedData = i.AuthenticatedData
	m1.RecursionAvailable = i.RecursionAvailable
	m1.Compress = true

	m1.Answer = i.Answer
	m1.Ns = i.Ns
	m1.Extra = i.Extra

	return m1
}

// nodataKey returns a caching key for NODATA responses.
func noDataKey(qname string, qtype uint16, do bool) string {
	if do {
		return "1" + qname + ".." + strconv.Itoa(int(qtype))
	}
	return "0" + qname + ".." + strconv.Itoa(int(qtype))
}

// nameErrorKey returns a caching key for NXDOMAIN responses.
func nameErrorKey(qname string, do bool) string {
	if do {
		return "1" + qname
	}
	return "0" + qname
}

// successKey returns a caching key for successfull answers.
func successKey(qname string, qtype uint16, do bool) string { return noDataKey(qname, qtype, do) }
