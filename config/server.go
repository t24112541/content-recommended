package config

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type ServerConfig struct {
	Prefork       bool   `env:"APP_PREFORK"`
	CaseSensitive bool   `env:"APP_CASE_SENSITIVE"`
	StrictRouting bool   `env:"APP_STRICT_ROUTING"`
	ServerHeader  string `env:"APP_SERVER_HEADER"`
	AppName       string `env:"APP_NAME"`
	AppPort       string `env:"APP_HTTP_PORT"`
}

func NewServer() (conf ServerConfig) {
	if err := env.Parse(&conf); err != nil {
		log.Fatalf("Failed Parse environment: %s", color.RedString(err.Error()))
	}

	return
}

func (conf ServerConfig) RunServer() (app *fiber.App, err error) {

	app = fiber.New(fiber.Config{
		Prefork:       conf.Prefork,
		CaseSensitive: conf.CaseSensitive,
		StrictRouting: conf.StrictRouting,
		ServerHeader:  conf.ServerHeader,
		AppName:       conf.AppName,
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
	}))

	return
}
