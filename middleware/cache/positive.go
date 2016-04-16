package cache

// successKey returns a caching key for successfull answers.
func successKey(qname string, qtype uint16, do bool) string { return noDataKey(qname, qtype, do) }
