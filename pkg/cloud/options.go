package cloud

// DiskOptions represents parameters to create an EBS volume.
type DiskOptions struct {
	CapacityBytes          int64
	Tags                   map[string]string
	VolumeType             string
	IOPSPerGB              int32
	AllowIOPSPerGBIncrease bool
	IOPS                   int32
	Throughput             int32
	AvailabilityZone       string
	OutpostArn             string
	Encrypted              bool
	BlockExpress           bool
	MultiAttachEnabled     bool
	// KmsKeyID represents a fully qualified resource name to the key to use for encryption.
	// example: arn:aws:kms:us-east-1:012345678910:key/abcd1234-a123-456a-a12b-a123b4cd56ef
	KmsKeyID   string
	SnapshotID string
}