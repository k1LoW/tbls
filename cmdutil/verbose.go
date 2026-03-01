package cmdutil

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"time"
)

var verbose bool
var startTime time.Time
var skipPartitions bool

// SetVerbose enables or disables verbose logging.
func SetVerbose(v bool) {
	verbose = v
	if v {
		startTime = time.Now()
	}
}

// IsVerbose returns whether verbose logging is enabled.
func IsVerbose() bool {
	return verbose
}

// Verbosef prints a verbose message with elapsed time if verbose mode is enabled.
func Verbosef(format string, args ...interface{}) {
	if !verbose {
		return
	}
	elapsed := time.Since(startTime)
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[%7.2fs] %s\n", elapsed.Seconds(), msg)
}

// SetSkipPartitions enables or disables skipping table partitions.
func SetSkipPartitions(v bool) {
	skipPartitions = v
}

// SkipPartitions returns whether to skip table partitions.
func SkipPartitions() bool {
	return skipPartitions
}

// MaskDSN masks the password in a DSN URL for safe logging.
func MaskDSN(dsn string) string {
	// Try to parse as URL first
	u, err := url.Parse(dsn)
	if err == nil && u.User != nil {
		if _, hasPassword := u.User.Password(); hasPassword {
			// Rebuild manually to avoid URL encoding of ****
			masked := u.Scheme + "://" + u.User.Username() + ":****@" + u.Host + u.Path
			if u.RawQuery != "" {
				masked += "?" + u.RawQuery
			}
			return masked
		}
		return dsn
	}

	// Fallback: use regex for various DSN formats
	// Matches patterns like user:password@ or password=secret
	re := regexp.MustCompile(`(://[^:]+:)[^@]+(@)`)
	masked := re.ReplaceAllString(dsn, "${1}****${2}")

	re2 := regexp.MustCompile(`(password=)[^\s&]+`)
	masked = re2.ReplaceAllString(masked, "${1}****")

	return masked
}
