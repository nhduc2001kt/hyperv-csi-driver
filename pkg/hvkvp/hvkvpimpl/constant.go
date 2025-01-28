package hvkvpimpl

import (
	"time"
)

const (
	// HyperVKVPConfigLoc is the path to the Hyper-V key value pairs configuration location.
	HyperVKVPConfigLoc = "/var/lib/hyperv"

	// HyperVKVPPoolCount is the number of key value pairs in the Hyper-V key value pairs pool.
	HyperVKVPPoolCount = 5

	// HyperVKVPPoolFilePrefix is the prefix for the Hyper-V key value pairs pool file.
	HyperVKVPPoolFilePrefix = ".kvp_pool_"

	// HyperVKVPPoolFileCheckInterval is the interval to check the Hyper-V key value pairs pool file.
	HyperVKVPPoolFileCheckInterval = 1 * time.Second

	// HyperVKVPPoolFileCheckTimeout is the timeout to check the Hyper-V key value pairs pool file.
	HyperVKVPPoolFileCheckTimeout = 5 * time.Minute

	// HyperVKVPEntriesPerBlock is the number of key value pairs entries per block.
	HyperVKVPEntriesPerBlock = 50

	// HyperVKPVExchangeMaxKeySize is the maximum key size for the Hyper-V key value pairs exchange.
	HyperVKPVExchangeMaxKeySize = 512

	// HyperVKPVExchangeMaxValueSize is the maximum value size for the Hyper-V key value pairs exchange.
	HyperVKPVExchangeMaxValueSize = 2048

	// HyperVKVPFileDescriptor is the path to the Hyper-V key value pairs file descriptor.
	HyperVKVPFileDescriptor = "/dev/vmbus/hv_kvp"

	// HyperVKVPOpRgister0 is the Hyper-V key value pairs operation register 0.
	HyperVKVPOpRgister0 = uint8(4)

	// HyperVKVPOpRgister1 is the Hyper-V key value pairs operation register 1.
	HyperVKVPOpRgister1 = uint8(100)

	// HyperVKVPPoolIn is the Hyper-V key value pairs pool in.
	HyperVKVPPoolIn = 0x0001

	// HyperVKVPPoolPri is the Hyper-V key value pairs pool priority.
	HyperVKVPPoolPri = 0x0002

	// HyperVKVPPoolOut is the Hyper-V key value pairs pool out.
	HyperVKVPPoolOut = 0x0004

	// HyperVKVPPoolErr is the Hyper-V key value pairs pool error.
	HyperVKVPPoolErr = 0x0008

	// HyperVKVPPoolUp is the Hyper-V key value pairs pool update.
	HyperVKVPPoolUp = 0x0010

	// HyperVKVPPoolVal is the Hyper-V key value pairs pool value.
	HyperVKVPPoolVal = 0x0020

	// HyperVKVPMsgBodySize is the size of the Hyper-V key value pairs message body.
	HyperVKVPMsgBodySize = 7428

	// HyperVKVPMessageSize is the size of the Hyper-V key value pairs message.
	HyperVKVPMessageSize = 7432

	// HyperVKVPNetDir is the Hyper-V key value pairs network directory.
	HyperVKVPNetDir = "/sys/class/net"

	// NetMaxAdapterIDSize is the maximum adapter ID size for the network adapter.
	NetMaxAdapterIDSize = 128

	// NetMaxIPAddrSize is the maximum IP address size for the network adapter.
	NetMaxIPAddrSize = 1024

	// NetMaxGatewaySize is the maximum gateway size for the network adapter.
	NetMaxGatewaySize = 512
)

const (
	OSReleaseFile     = "/etc/os-release"
	SuSEReleaseFile   = "/etc/SuSE-release"
	RedHatReleaseFile = "/etc/redhat-release"
)

const (
	// HyperVKVPOpGet is the Hyper-V key value pairs operation get.
	HyperVKVPOpGet uint8 = iota
	// HyperVKVPOpSet is the Hyper-V key value pairs operation set.
	HyperVKVPOpSet
	// HyperVKVPOpDelete is the Hyper-V key value pairs operation delete.
	HyperVKVPOpDelete
	// HyperVKVPOpEnumerate is the Hyper-V key value pairs operation enumerate.
	HyperVKVPOpEnumerate
	// HyperVKVPOpGetIPInfo is the Hyper-V key value pairs operation get IP information.
	HyperVKVPOpGetIPInfo
	// HyperVKVPOpSetIPInfo is the Hyper-V key value pairs operation set IP information.
	HyperVKVPOpSetIPInfo
	// HyperVKVPOpCount is the Hyper-V key value pairs operation count.
	HyperVKVPOpCount
)
