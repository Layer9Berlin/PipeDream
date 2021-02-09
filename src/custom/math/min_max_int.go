// Package math contains custom functions extending the native `math` package
package math

// MaxInt returns the larger of two integers
func MaxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt returns the smaller of two integers
func MinInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
