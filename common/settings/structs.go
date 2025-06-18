package settings

type AllowedMethods string

const (
	ProxyAuth    AllowedMethods = "proxyAuth"
	NoAuth       AllowedMethods = "noAuth"
	PasswordAuth AllowedMethods = "passwordAuth"
)

type Settings struct {
	Server       Server       `json:"server"`
	Auth         Auth         `json:"auth"`
	Frontend     Frontend     `json:"frontend"`
	UserDefaults UserDefaults `json:"userDefaults"`
	Integrations Integrations `json:"integrations"`
}

type LogConfig struct {
	Levels    string `json:"levels"`    // separated list of log levels to enable. (eg. "info|warning|error|debug")
	ApiLevels string `json:"apiLevels"` // separated list of log levels to enable for the API. (eg. "info|warning|error")
	Output    string `json:"output"`    // output location. (eg. "stdout" or "path/to/file.log")
	NoColors  bool   `json:"noColors"`  // disable colors in the output
	Json      bool   `json:"json"`      // output in json format, currently not supported
	Utc       bool   `json:"utc"`       // use UTC time in the output instead of local time
}

type Server struct {
	NumImageProcessors           int         `json:"numImageProcessors"`           // number of concurrent image processing jobs used to create previews, default is number of cpu cores available.
	Socket                       string      `json:"socket"`                       // socket to listen on
	TLSKey                       string      `json:"tlsKey"`                       // path to TLS key
	TLSCert                      string      `json:"tlsCert"`                      // path to TLS cert
	DisablePreviews              bool        `json:"disablePreviews"`              // disable all previews thumbnails, simple icons will be used
	DisableResize                bool        `json:"disablePreviewResize"`         // disable resizing of previews for faster loading over slow connections
	DisableTypeDetectionByHeader bool        `json:"disableTypeDetectionByHeader"` // disable type detection by header, useful if filesystem is slow.
	Port                         int         `json:"port"`                         // port to listen on
	BaseURL                      string      `json:"baseURL"`                      // base URL for the server, the subpath that the server is running on.
	Logging                      []LogConfig `json:"logging"`
	DebugMedia                   bool        `json:"debugMedia"` // output ffmpeg stdout for media integration -- careful can produces lots of output!
	Database                     string      `json:"database"`   // path to the database file
	Sources                      []Source    `json:"sources" validate:"required,dive"`
	ExternalUrl                  string      `json:"externalUrl"`    // used by share links if set
	InternalUrl                  string      `json:"internalUrl"`    // used by integrations if set, this is the url that an integration service will use to communicate with filebrowser
	CacheDir                     string      `json:"cacheDir"`       // path to the cache directory, used for thumbnails and other cached files
	MaxArchiveSizeGB             int64       `json:"maxArchiveSize"` // max pre-archive combined size of files/folder that are allowed to be archived (in GB)
	// not exposed to config
	SourceMap      map[string]Source `json:"-" validate:"omitempty"` // uses realpath as key
	NameToSource   map[string]Source `json:"-" validate:"omitempty"` // uses name as key
	DefaultSource  Source            `json:"-" validate:"omitempty"`
	MuPdfAvailable bool              `json:"-"` // used internally if compiled with mupdf support
}

type Source struct {
	Path   string       `json:"path" validate:"required"` // file system path. (Can be relative)
	Name   string       `json:"name"`                     // display name
	Config SourceConfig `json:"config"`
}

type SourceConfig struct {
	IndexingInterval      uint32      `json:"indexingIntervalMinutes"` // optional manual overide interval in seconds to re-index the source
	DisableIndexing       bool        `json:"disableIndexing"`         // disable the indexing of this source
	MaxWatchers           int         `json:"maxWatchers"`             // number of concurrent watchers to use for this source, currently not supported
	NeverWatch            []string    `json:"neverWatchPaths"`         // paths to never watch, relative to the source path (eg. "/folder/file.txt")
	IgnoreHidden          bool        `json:"ignoreHidden"`            // ignore hidden files and folders.
	IgnoreZeroSizeFolders bool        `json:"ignoreZeroSizeFolders"`   // ignore folders with 0 size
	Exclude               IndexFilter `json:"exclude"`                 // exclude files and folders from indexing, if include is not set
	Include               IndexFilter `json:"include"`                 // include files and folders from indexing, if exclude is not set
	DefaultUserScope      string      `json:"defaultUserScope"`        // default "/" should match folders under path
	DefaultEnabled        bool        `json:"defaultEnabled"`          // should be added as a default source for new users?
	CreateUserDir         bool        `json:"createUserDir"`           // create a user directory for each user
}
type IndexFilter struct {
	Files        []string `json:"files"`        // array of file names to include/exclude
	Folders      []string `json:"folders"`      // array of folder names to include/exclude
	FileEndsWith []string `json:"fileEndsWith"` // array of file names to include/exclude (eg "a.jpg")
}
