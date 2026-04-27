package cmdutil

var skipPartitions bool

// SetSkipPartitions enables or disables skipping table partitions.
func SetSkipPartitions(v bool) {
	skipPartitions = v
}

// SkipPartitions returns whether to skip table partitions.
func SkipPartitions() bool {
	return skipPartitions
}
