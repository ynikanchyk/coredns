package dns

import (
	"path/filepath"

	"github.com/miekg/coredns/core/assets"
)

// storage is used to get file paths in a consistent,
// cross-platform way for persisting Let's Encrypt assets
// on the file system.
var storage = Storage(filepath.Join(assets.Path(), "letsencrypt"))

// Storage is a root directory and facilitates
// forming file paths derived from it.
type Storage string

// Zones gets the directory that stores zone keys.
func (s Storage) Zones() string {
	return filepath.Join(string(s), "zones")
}

// Zone returns the path to the folder containing assets for domain.
func (s Storage) Site(domain string) string {
	return filepath.Join(s.Zones(), domain)
}

// SecondaryZoneFile
// SiteKeyFile returns the path to domain's private key file.
func (s Storage) SiteKeyFile(domain string) string {
	return filepath.Join(s.Site(domain), domain+".key")
}

// Expand will expand name:
// - If name starts with a / it is absolute and taken as is.
// - if name starts with ~/ it point to assets.Path(), i.e CoreDNS' homedirectory
// - If a name does not start with the above characters it is taken as relative
//      to home and an absolute path is created from that.
func (s Storage) Expand(home, name string) string {

}
