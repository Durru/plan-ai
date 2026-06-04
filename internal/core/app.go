package core

type App struct {
	Version string
	RootDir string
}

func NewApp(version, rootDir string) App {
	return App{Version: version, RootDir: rootDir}
}
