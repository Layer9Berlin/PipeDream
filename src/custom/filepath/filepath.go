// Package filepath contains custom functions extending the native `filepath` package
package filepath

import (
	systemfilepath "path/filepath"
)

// AbsolutePath checks if path is absolute and if not, prepends the basePath
func AbsolutePath(path string, basePath string) (string, error) {
	if systemfilepath.IsAbs(path) {
		return path, nil
	}
	return systemfilepath.Join(basePath, path), nil
}
