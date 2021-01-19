package custom_filepath

import (
	"path/filepath"
)

func AbsolutePath(path string, basePath string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Join(basePath, path), nil
}
