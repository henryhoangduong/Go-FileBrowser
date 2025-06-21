package iteminfo

import (
	"path/filepath"
	"time"
)

type ItemInfo struct {
	Name    string    `json:"name"`     // name of the file
	Size    int64     `json:"size"`     // length in bytes for regular files
	ModTime time.Time `json:"modified"` // modification time
	Type    string    `json:"type"`     // type of the file, either "directory" or a file mimetype
	Hidden  bool      `json:"hidden"`   // whether the file is hidden
}
type FileInfo struct {
	ItemInfo
	Files   []ItemInfo `json:"files"`   // files in the directory
	Folders []ItemInfo `json:"folders"` // folders in the directory
	Path    string     `json:"path"`    // path scoped to the associated index
}
type ExtendedFileInfo struct {
	FileInfo
	Content      string            `json:"content,omitempty"`      // text content of a file, if requested
	Subtitles    []string          `json:"subtitles,omitempty"`    // subtitles for video files
	Checksums    map[string]string `json:"checksums,omitempty"`    // checksums for the file
	Token        string            `json:"token,omitempty"`        // token for the file -- used for sharing
	OnlyOfficeId string            `json:"onlyOfficeId,omitempty"` // id for onlyoffice files
	Source       string            `json:"source"`                 // associated index source for the file
	RealPath     string            `json:"-"`
}
type FileOptions struct {
	Path       string // realpath
	Source     string
	IsDir      bool
	Modify     bool
	Expand     bool
	ReadHeader bool
	Content    bool
}

func (f FileOptions) Components() (string, string) {
	return filepath.Dir(f.Path), filepath.Base(f.Path)
}
