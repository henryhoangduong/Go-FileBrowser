package cmd

import (
	"filebrowser/common/settings"
	"os"

	"gorm.io/gorm/logger"
)

var (
	configPath string
)

func runCLI() bool {
	generateYaml()

}
func generateYaml() {
	if os.Getenv("FILEBROWSER_GENERATE_CONFIG") != "" {
		logger.Info("Generating config.yaml")
		settings.GenerateYaml()
		os.Exit(0)
	}
}
