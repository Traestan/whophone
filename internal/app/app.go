package app

import (
	"github.com/traestan/whophone/internal/config"
	"github.com/traestan/whophone/internal/db"
	"github.com/traestan/whophone/internal/numcap"
	"go.uber.org/zap"
)

type WhPhone interface {
	Find(string) (bool, error)
	UpdateBase() (bool, error)
}
type whoPhone struct {
	config config.TomlConfig
	logger *zap.Logger
	db     *db.Client
}

func NewWhoPhone(cfg config.TomlConfig, logger *zap.Logger) (WhPhone, error) {
	mng, err := db.NewClient(&cfg.Db, logger)
	if err != nil {
		return nil, err
	}
	return &whoPhone{
		config: cfg,
		logger: logger,
		db:     mng,
	}, nil
}

// Find phone search
func (app *whoPhone) Find(phone string) (bool, error) {
	app.logger.Info("Start find phone number", zap.String("number", phone))
	findInfo := app.db.GetHash(phone)
	if findInfo.Operator == "" {
		app.logger.Info("The number is not found", zap.String("number", phone))
		app.db.Close()
		return false, nil
	}
	app.logger.Info("The number is found",
		zap.String("number", phone),
		zap.String("operator", findInfo.Operator),
		zap.String("region", findInfo.Region))
	app.db.Close()

	return true, nil
}

// UpdateBase database update
func (app *whoPhone) UpdateBase() (bool, error) {
	app.db.StatsInfo()
	app.logger.Info("Start update base")
	nc, err := numcap.NewNumCap(app.db, app.logger, app.config)
	if err != nil {
		return false, err
	}
	stateDownload, err := nc.Download()
	if err != nil {
		return false, err
	}
	if stateDownload { // files downloaded, need parsing
		// clear old info
		_, err := app.db.ClearOld()
		if err != nil {
			return false, err
		}

		stateAddStorage, err := nc.Csv2Storage()
		if err != nil {
			app.logger.Panic(err.Error())
			return false, err
		}
		if stateAddStorage {
			app.db.StatsInfo()
		}
	}
	app.logger.Info("End update")
	app.db.Close()
	return true, nil
}
