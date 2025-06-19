package settings

import (
	"filebrowser/common/version"
	"filebrowser/database/users"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/gtsteffaniak/go-logger/logger"
)

var Config Settings

const (
	generatorPath = "/relative/or/absolute/path"
)

func Initialize(configFile string) {}

func setupLogging() {
	if len(Config.Server.Logging) == 0 {
		Config.Server.Logging = []LogConfig{
			{
				Output: "stdout",
			},
		}
	}
	for _, logConfig := range Config.Server.Logging {
		logConfig := logger.JsonConfig{
			Levels:    logConfig.Levels,
			ApiLevels: logConfig.ApiLevels,
			Output:    logConfig.Output,
			Utc:       logConfig.Utc,
			NoColors:  logConfig.NoColors,
			Json:      logConfig.Json,
		}
		err := logger.SetupLogger(logConfig)
		if err != nil {
			log.Println("[ERROR] Failed to set up logger:", err)
		}
	}
}
func loadConfigWithDefaults(configFile string) error {
	Config = setDefaults()
	// Open and read the YAML file
	yamlFile, err := os.Open(configFile)
	if err != nil {
		logger.Errorf("could not open config file '%v', using default settings.", configFile)
		Config.Server.Sources = []Source{
			{
				Path: ".",
			},
		}
		loadEnvConfig()
		return nil
	}
	defer yamlFile.Close()
	stat, err := yamlFile.Stat()
	if err != nil {
		return err
	}
	yamlData := make([]byte, stat.Size())
	_, err = yamlFile.Read(yamlData)
	if err != nil && configFile != "config.yaml" {
		return fmt.Errorf("could not load specified config file: %v", err.Error())
	}
	if err != nil {
		logger.Warningf("Could not load config file '%v', using default settings: %v", configFile, err)
	}
	err = yaml.NewDecoder(strings.NewReader(string(yamlData)), yaml.DisallowUnknownField()).Decode(&Config)
	if err != nil {
		return fmt.Errorf("error unmarshaling YAML data: %v", err)
	}
	loadEnvConfig()
	return nil
}
func setDefaults() Settings {
	numCpus := 4
	if cpus := runtime.NumCPU(); cpus > 0 {
		numCpus = cpus
	}
	database := os.Getenv("FILEBROWSER_DATABASE")
	if database == "" {
		database = "database.db"
	}
	return Settings{
		Server: Server{
			Port:               80,
			NumImageProcessors: numCpus,
			BaseURL:            "",
			Database:           database,
			SourceMap:          map[string]Source{},
			NameToSource:       map[string]Source{},
			MaxArchiveSizeGB:   50,
			CacheDir:           "tmp",
		}, Auth: Auth{
			AdminUsername:        "admin",
			AdminPassword:        "admin",
			TokenExpirationHours: 2,
			Methods: LoginMethods{
				PasswordAuth: PasswordAuthConfig{
					Enabled:   true,
					MinLength: 5,
					Signup:    false,
				},
			},
		}, Frontend: Frontend{
			Name: "FileBrowser Quantum",
		}, UserDefaults: UserDefaults{
			DisableOnlyOfficeExt: ".txt .csv .html .pdf",
			StickySidebar:        true,
			LockPassword:         false,
			ShowHidden:           false,
			DarkMode:             true,
			DisableSettings:      false,
			ViewMode:             "normal",
			Locale:               "en",
			GallerySize:          3,
			ThemeColor:           "var(--blue)",
			Permissions: users.Permissions{
				Modify: false,
				Share:  false,
				Admin:  false,
				Api:    false,
			},
		},
	}
}

func loadEnvConfig() {
	adminPassword, ok := os.LookupEnv("FILEBROWSER_ADMIN_PASSWORD")
	if ok {
		logger.Info("Using admin password from FILEBROWSER_ADMIN_PASSWORD environment variable")
		Config.Auth.AdminPassword = adminPassword
	}
	officeSecret, ok := os.LookupEnv("FILEBROWSER_ONLYOFFICE_SECRET")
	if ok {
		logger.Info("Using OnlyOffice secret from FILEBROWSER_ONLYOFFICE_SECRET environment variable")
		Config.Integrations.OnlyOffice.Secret = officeSecret
	}
	ffmpegPath, ok := os.LookupEnv("FILEBROWSER_FFMPEG_PATH")
	if ok {
		Config.Integrations.Media.FfmpegPath = ffmpegPath
	}
}
func setupAuth() {
	if Config.Auth.Methods.PasswordAuth.Enabled {
		Config.Auth.AuthMethods = append(Config.Auth.AuthMethods, "password")
	}
	if Config.Auth.Methods.ProxyAuth.Enabled {
		Config.Auth.AuthMethods = append(Config.Auth.AuthMethods, "proxy")
	}
	if Config.Auth.Methods.OidcAuth.Enabled {
		Config.Auth.AuthMethods = append(Config.Auth.AuthMethods, "oidc")
	}
	if Config.Auth.Methods.NoAuth {
		logger.Warning("Configured with no authentication, this is not recommended.")
		Config.Auth.AuthMethods = []string{"disabled"}
	}
	if Config.Auth.Methods.OidcAuth.Enabled {
		err := validateOidcAuth()
		if err != nil {
			logger.Fatalf("Error validating OIDC auth: %v", err)
		}
		logger.Info("OIDC Auth configured successfully")
	}
	if len(Config.Auth.AuthMethods) == 0 {
		Config.Auth.Methods.PasswordAuth.Enabled = true
		Config.Auth.AuthMethods = append(Config.Auth.AuthMethods, "password")
	}
	Config.UserDefaults.LoginMethod = Config.Auth.AuthMethods[0]
}

func setupSources(generate bool) {
	if len(Config.Server.Sources) == 0 {
		logger.Fatal("There are no `server.sources` configured. If you have `server.root` configured, please update the config and add at least one `server.sources` with a `path` configured.")
	} else {
		for k, source := range Config.Server.Sources {
			realPath := getRealPath(source.Path)
			name := filepath.Base(realPath)
			if name == "\\" {
				name = strings.Split(realPath, ":")[0]
			}
			if generate {
				source.Path = generatorPath // use placeholder path
			} else {
				source.Path = realPath // use absolute path
			}
			if source.Name == "" {
				_, ok := Config.Server.SourceMap[source.Path]
				if ok {
					source.Name = name + fmt.Sprintf("-%v", k)
				} else {
					source.Name = name
				}
				if Config.Server.DefaultSource.Path == "" {
					Config.Server.DefaultSource = source
				}
			}
			if source.Config.DefaultUserScope == "" {
				source.Config.DefaultUserScope = "/"
			}
			Config.Server.SourceMap[source.Path] = source
			Config.Server.NameToSource[source.Name] = source
		}
	}
	sourceList := []Source{}
	defaultScopes := []users.SourceScope{}
	allSourceNames := []string{}
	first := true
	potentialDefaultSource := Config.Server.DefaultSource
	for _, sourcePathOnly := range Config.Server.Sources {
		realPath := getRealPath(sourcePathOnly.Path)
		if generate {
			realPath = generatorPath // use placeholder path
		}
		source, ok := Config.Server.SourceMap[realPath]
		if ok && !slices.Contains(allSourceNames, source.Name) {
			if first {
				potentialDefaultSource = source
			}
			first = false
			sourceList = append(sourceList, source)
			if source.Config.DefaultEnabled {
				if Config.Server.DefaultSource.Path == "" {
					Config.Server.DefaultSource = source
				}
				defaultScopes = append(defaultScopes, users.SourceScope{
					Name:  source.Path,
					Scope: source.Config.DefaultUserScope,
				})
			}
			allSourceNames = append(allSourceNames, source.Name)
		} else {
			logger.Warningf("source %v is not configured correctly, skipping", sourcePathOnly.Path)
		}
	}
	if Config.Server.DefaultSource.Path == "" {
		Config.Server.DefaultSource = potentialDefaultSource
	}
	sourceList2 := []Source{}
	for _, s := range sourceList {
		if s.Path == Config.Server.DefaultSource.Path {
			s.Config.DefaultEnabled = true
			Config.Server.SourceMap[s.Path] = s
			Config.Server.NameToSource[s.Name] = s
		}
		sourceList2 = append(sourceList2, s)
	}
	sourceList = sourceList2
	Config.UserDefaults.DefaultScopes = defaultScopes
	Config.Server.Sources = sourceList
}
func getRealPath(path string) string {
	realPath, err := filepath.Abs(path)
	if err != nil {
		logger.Fatalf("could not find configured source path: %v", err)
	}
	// check path exists
	if _, err = os.Stat(realPath); os.IsNotExist(err) {
		logger.Fatalf("configured source path does not exist: %v", realPath)
	}
	return realPath
}
func setupUrls() {
	baseurl := strings.Trim(Config.Server.BaseURL, "/")
	if baseurl == "" {
		Config.Server.BaseURL = "/"
	} else {
		Config.Server.BaseURL = "/" + baseurl + "/"
	}
	Config.Server.InternalUrl = strings.Trim(Config.Server.InternalUrl, "/")
	Config.Server.ExternalUrl = strings.Trim(Config.Server.ExternalUrl, "/")
	Config.Integrations.OnlyOffice.Url = strings.Trim(Config.Integrations.OnlyOffice.Url, "/")
	Config.Integrations.OnlyOffice.InternalUrl = strings.Trim(Config.Integrations.OnlyOffice.InternalUrl, "/")
}
func setupFrontend() {
	if !Config.Frontend.DisableDefaultLinks {
		Config.Frontend.ExternalLinks = append(Config.Frontend.ExternalLinks, ExternalLink{
			Text:  fmt.Sprintf("(%v)", version.Version),
			Title: version.CommitSHA,
			Url:   "https://github.com/gtsteffaniak/filebrowser/releases/",
		})
		Config.Frontend.ExternalLinks = append(Config.Frontend.ExternalLinks, ExternalLink{
			Text: "Help",
			Url:  "https://github.com/gtsteffaniak/filebrowser/wiki",
		})
	}
}
