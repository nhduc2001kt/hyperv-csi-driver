package addressfamily

type AddressFamily int8

const (
	AddressFamilyNone AddressFamily = iota
	AddressFamilyIPv4
	AddressFamilyIPv6
)
