package networkblocker

type API interface {
	// RejectTCP actively denies all traffic on the given TCP port
	RejectTCP(port int) error

	// DropTCP silently denies all traffic on the given TCP port
	DropTCP(port int) error

	// AcceptTCP allow all traffic on the given TCP port
	AcceptTCP(port int) error

	// Rules returns a list of all rules injected by this service.
	Rules() ([]string, error)
}
