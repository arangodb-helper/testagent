package networkblocker

type API interface {
	// RejectTCP actively denies all traffic on the given TCP port
	RejectTCP(port int) error

	// DropTCP silently denies all traffic on the given TCP port
	DropTCP(port int) error

	// AcceptTCP allow all traffic on the given TCP port
	AcceptTCP(port int) error

	// RejectAllFrom actively denies all traffic coming from the given IP address on the given interface
	RejectAllFrom(ip, intf string) error

	// DropAllFrom silently denies all traffic coming from the given IP address on the given interface
	DropAllFrom(ip, intf string) error

	// AcceptAllFrom allow all traffic coming from the given IP address on the given interface
	AcceptAllFrom(ip, intf string) error

	// Rules returns a list of all rules injected by this service.
	Rules() ([]string, error)
}
