// Custom functions extending the native `filepath` package
package filepath

import (
	systemfilepath "path/filepath"
)

func AbsolutePath(path string, basePath string) (string, error) {
	if systemfilepath.IsAbs(path) {
		return path, nil
	}
	return systemfilepath.Join(basePath, path), nil
}
