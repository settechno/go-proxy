package application

import (
	"log"
	"proxy/internal/config"
	"proxy/internal/storage/users"
	"proxy/internal/storage/users/json"
)

type App struct {
	Config      *config.AppConfig
	UserStorage users.UserStorageInterface
}

func (app *App) init() {
	// Загрузка конфигурации
	cfg, err := config.LoadAppConfig(config.GetAppConfigFile())
	if err != nil {
		log.Fatal(err)
	}
	app.Config = cfg

	// Загрузка пользователей
	app.UserStorage = json.NewStorage(cfg)
}

func NewApp() *App {
	app := &App{}
	app.init()

	return app
}
