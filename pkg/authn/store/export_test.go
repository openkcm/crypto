package store

import "path/filepath"

// NewFSWithDir creates a new filesystem store that persists tokens to the specified directory.
func NewFSWithDir(dir string) *FS {
	return &FS{
		dirPath:   dir,
		tokenPath: filepath.Join(dir, TokenFileName),
	}
}
