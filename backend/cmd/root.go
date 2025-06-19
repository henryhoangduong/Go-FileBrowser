package cmd

import (
	"filebrowser/common/settings"
	"fmt"

	"github.com/gtsteffaniak/go-logger/logger"
)

func generalUsage() {
	fmt.Printf(`usage: ./filebrowser <command> [options]
commands:
	-h    	Print help
	-c    	Print the default config file
	version Print version information
	set -u	Username and password for the new user
	set -a	Create user as admin
	set -s	Specify a user scope
	set -h	Print this help message
`)
}
func getStore(configFile string) bool {
	// Use the config file (global flag)
	settings.Initialize(configFile)
	s, hasDB, err := storage.InitializeDb(settings.Config.Server.Database)
	if err != nil {
		logger.Fatalf("could not load db info: %v", err)
	}
	store = s
	return hasDB
}
