package bloxpolicy


// IsClientAllowed returns true if DNS lookups are permitted from the client
// based on the blox policy engine.
func (p BloxPolicy) IsClientAllowed(client string) (bool, error) {

	// Initially hard-corded
	allowed := true

	var err error 

	
	// TODO: JB integration here


	return allowed, err
}
