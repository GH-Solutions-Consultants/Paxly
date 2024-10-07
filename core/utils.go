// core/utils.go
package core

import "runtime"

// IsWindows checks if the current OS is Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}