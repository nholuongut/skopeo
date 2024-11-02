// +build !windows

package system

// LCOWSupported returns true if Linux nholuongut on Windows are supported.
func LCOWSupported() bool {
	return false
}
