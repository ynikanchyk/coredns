package cache

import "github.com/miekg/dns"

type Item struct {
	Authoritative      bool
	AuthenticatedData  bool
	RecursionAvailable bool
	Answer             []dns.RR
	Ns                 []dns.RR
	Extra              []dns.RR
}

func newItem(m *dns.Msg) *Item {
	i := new(Item)
	i.Authoritative = m.Authoritative
	i.AuthenticatedData = m.AuthenticatedData
	i.RecursionAvailable = m.RecursionAvailable
	i.Answer = m.Answer
	i.Ns = m.Ns
	i.Extra = m.Extra

	return i
}

// toMsg turns i into a message, it tailers to reply to m.
func (i *Item) toMsg(m *dns.Msg) *dns.Msg {
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
