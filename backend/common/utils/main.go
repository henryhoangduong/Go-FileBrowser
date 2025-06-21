package utils

import (
	"filebrowser/indexing/iteminfo"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolveSymlinks(path string) (string, bool, error) {
	for {
		// Get the file info using os.Lstat to handle symlinks
		info, err := os.Lstat(path)
		if err != nil {
			return path, false, fmt.Errorf("could not stat path: %s, %v", path, err)
		}

		// Check if the path is a symlink
		if info.Mode()&os.ModeSymlink != 0 {
			// Read the symlink target
			target, err := os.Readlink(path)
			if err != nil {
				return path, false, fmt.Errorf("could not read symlink: %s, %v", path, err)
			}

			// Resolve the symlink's target relative to its directory
			path = filepath.Join(filepath.Dir(path), target)
		} else {
			// Not a symlink, check with bundle-aware directory logic
			isDir := iteminfo.IsDirectory(info)
			return path, isDir, nil
		}
	}
}
func GetParentDirectoryPath(path string) string {
	if path == "/" || path == "" {
		return ""
	}
	path = strings.TrimSuffix(path, "/") // Remove trailing slash if any
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return "" // No parent directory for a relative path without slashes
	}
	if lastSlash == 0 {
		return "/" // If the last slash is the first character, return root
	}
	return path[:lastSlash]
}
