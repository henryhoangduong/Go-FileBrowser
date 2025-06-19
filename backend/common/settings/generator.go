package settings

type commentsMap map[string]map[string]string

func GenerateYaml() {
	_ = loadConfigWithDefaults("")
	Config.Server.Sources = []Source{
		{
			Path: ".",
		},
	}
	setupLogging()
	setupAuth()
	setupSources(true)
	setupUrls()

}
