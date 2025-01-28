package keyindex

type KeyIndex int8

const (
	FullyQualifiedDomainName KeyIndex = iota
	IntegrationServicesVersion
	NetworkAddressIPv4
	NetworkAddressIPv6
	OSBuildNumber
	OSName
	OSMajorVersion
	OSMinorVersion
	OSVersion
	ProcessorArchitecture
)

