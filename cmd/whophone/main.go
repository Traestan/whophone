package main

import (
	"flag"

	"github.com/BurntSushi/toml"
	"github.com/traestan/whophone/internal/app"
	"github.com/traestan/whophone/internal/config"
	"go.uber.org/zap"
)

func main() {
	var (
		configFile  = flag.String("config.toml", "../../config.toml", "Service port")
		config      config.TomlConfig
		phoneNumber = flag.String("phone", "", "The phone number you need to find")
		phoneAction = flag.String("action", "", "Action update, find")
	)
	flag.Parse()

	if _, err := toml.DecodeFile(*configFile, &config); err != nil {
		panic(err)
	}
	// logger
	logger := zap.NewExample()
	defer logger.Sync()
	logger.Info("Who phone start")

	app, err := app.NewWhoPhone(config, logger)
	if err != nil {
		panic(err)
	}
	// Handles
	if *phoneAction == "update" {
		state, err := app.UpdateBase()
		if err != nil {
			panic(err)
		}
		logger.Info("State ", zap.Bool("update base ", state))
	}
	if *phoneAction == "find" && *phoneNumber != "" {
		_, err := app.Find(*phoneNumber)
		if err != nil {
			panic(err)
		}
	}
}
