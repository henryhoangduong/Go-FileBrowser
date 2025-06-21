package indexing

import (
	"filebrowser/common/settings"
	"filebrowser/common/utils"
	"filebrowser/indexing/iteminfo"
	"fmt"
	"path/filepath"
	"sync"
	"github.com/shirou/gopsutil/v3/disk"

	"github.com/gtsteffaniak/go-logger/logger"
)

func GetIndex(name string) *Index {
	indexesMutex.Lock()
	defer indexesMutex.Unlock()
	index, ok := indexes[name]
	if !ok {
		// try path if name fails
		// todo: update everywhere else so this isn't needed.
		source, ok := settings.Config.Server.SourceMap[name]
		if !ok {
			logger.Errorf("index %s not found", name)
		}
		index, ok = indexes[source.Name]
		if !ok {
			logger.Errorf("index %s not found", name)
		}

	}
	return index
}

var (
	indexes      map[string]*Index
	indexesMutex sync.RWMutex
)

func (idx *Index) UpdateMetadata(info *iteminfo.FileInfo) bool {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.Directories[info.Path] = info
	return true
}
func (idx *Index) GetReducedMetadata(target string, isDir bool) (*iteminfo.FileInfo, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	checkDir := idx.MakeIndexPath(target)
	if !isDir {
		checkDir = idx.MakeIndexPath(filepath.Dir(target))
	}
	if checkDir == "" {
		checkDir = "/"
	}
	dir, exists := idx.Directories[checkDir]
	if !exists {
		return nil, false
	}

	if isDir {
		return dir, true
	}
	// handle file
	if checkDir == "/" {
		checkDir = ""
	}
	baseName := filepath.Base(target)
	for _, item := range dir.Files {
		if item.Name == baseName {
			return &iteminfo.FileInfo{
				Path:     checkDir + "/" + item.Name,
				ItemInfo: item,
			}, true
		}
	}
	return nil, false

}
func (idx *Index) GetMetadataInfo(target string, isDir bool) (*iteminfo.FileInfo, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	checkDir := idx.MakeIndexPath(target)
	if !isDir {
		checkDir = idx.MakeIndexPath(filepath.Dir(target))
	}
	if checkDir == "" {
		checkDir = "/"
	}
	dir, exists := idx.Directories[checkDir]
	return dir, exists
}
func GetIndexInfo(sourceName string) (ReducedIndex, error) {
	idx, ok := indexes[sourceName]
	if !ok {
		return ReducedIndex{}, fmt.Errorf("index %s not found", sourceName)
	}
	sourcePath := idx.Path
	cacheKey := "usageCache-" + sourceName
	_, ok = utils.DiskUsageCache.Get(cacheKey).(bool)
	if !ok {
		usage, err := disk.Usage(sourcePath)
		if err != nil {
			logger.Errorf("error getting disk usage for %s: %v", sourcePath, err)
			idx.SetStatus(UNAVAILABLE)
			return ReducedIndex{}, fmt.Errorf("error getting disk usage for %s: %v", sourcePath, err)
		}
		latestUsage := DiskUsage{
			Total: usage.Total,
			Used:  usage.Used,
		}
		idx.SetUsage(latestUsage)
		utils.DiskUsageCache.Set(cacheKey, true)
	}
	return idx.ReducedIndex, nil
}
