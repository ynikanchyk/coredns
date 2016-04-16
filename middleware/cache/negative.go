package cache

import "strconv"

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
