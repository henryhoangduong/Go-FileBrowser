package settings

import (
	"filebrowser/database/users"
	"fmt"
	"log"
	"os"
	"runtime"
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
