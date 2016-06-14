package bloxpolicy

import (
	"fmt"
	"os"
)

// blockFilePolicy returns true if the file "block" exists in the pwd beside coredns
// Note: this function exists to simulate calling an external policy engine
func blockFilePolicy() (bool, error) {

	allow := false

	_, err := os.Stat("./block")
	if err != nil {
		if os.IsNotExist(err) {
			// file "./block" does not exist, therefore client is allowed, and
			// err has been handled.
			allow = true
			err = nil
		} else {
			// some other error occured, therefore default to disallow
			allow = false
		}
	} else {
		allow = false
		err = nil
	}
	
	return allow, err
}


// IsClientAllowed returns true if DNS lookups are permitted from the client
// based on the blox policy engine.
func (p BloxPolicy) IsClientAllowed(client string) (bool, error) {

	allow, err := blockFilePolicy()
	// TODO: JB integration here instead of call to blockFilePolicy

	fmt.Printf("#### exit IsClientAllowed: %v,%v\n", allow, err)
	return allow, err
}
