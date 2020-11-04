package main

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type MigrateConfig struct {
	DatabaseUrl   string /*`env:"DB_URL" envDefault:"postgres://postgres:devtronpg@127.0.0.1:5432/migrate_test?sslmode=disable"`*/
	TargetVersion uint   `env:"MIGRATE_TO_VERSION" envDefault:"0"`
	DbType        string `env:"DB_TYPE"  envDefault:"postgres"`
	UserName      string `env:"DB_USER_NAME"  envDefault:"postgres"`
	Password      string `env:"DB_PASSWORD"  `
	Host          string `env:"DB_HOST"  envDefault:"localhost"`
	Port          string `env:"DB_PORT"  envDefault:"5432"`
	DbName        string `env:"DB_NAME"  envDefault:"migrate_test"`
}


func (cfg MigrateConfig) Valid() bool {
	if cfg.DatabaseUrl == "" {
		return false
	} else {
		return true
	}
}

func GetMigrateConfig() (*MigrateConfig, error) {
	cfg := &MigrateConfig{}
	err := env.Parse(cfg)
	url := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", cfg.DbType, cfg.UserName, cfg.Password, cfg.Host, cfg.Port, cfg.DbName)
	cfg.DatabaseUrl = url
	return cfg, err
}

type MigrateUtil struct {
	config *MigrateConfig
	logger *zap.SugaredLogger
}

func NewMigrateUtil(cfg *MigrateConfig, logger *zap.SugaredLogger) *MigrateUtil {
	return &MigrateUtil{config: cfg, logger: logger}
}

func (util MigrateUtil) Migrate(sourceLocation string) (version uint, err error) {
	if util.config.TargetVersion != 0 {
		return util.migrateGoTo(fmt.Sprintf("file://%s", sourceLocation), util.config.TargetVersion)
	} else {
		return util.migrateUp(fmt.Sprintf("file://%s", sourceLocation))
	}
}

func (util MigrateUtil) migrateUp(location string) (version uint, err error) {
	util.logger.Infow("migrating from", "location", location)
	m, err := migrate.New(location, util.config.DatabaseUrl)
	if err != nil {
		util.logger.Errorw("error in connection db", "err", err)
		return 0, err
	}
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		util.logger.Errorw("error in determining current migration will not apply any further migration ", "err", err)
		return 0, err
	} else if err == migrate.ErrNilVersion {
		util.logger.Infow("applying changes on fresh db")
	} else if dirty {
		util.logger.Errorw("dirty db state found will not apply changes. please clean db manually first", "CurrentVersion", version)
		return 0, fmt.Errorf("dirty db found")
	} else {
		util.logger.Infow("applying changes over version ", "version", version)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			util.logger.Info("no changes detected", "version", version)
			return version, nil
		}
		util.logger.Errorw("error in applying migration", "err", err)
		return 0, err
	}
	version, dirty, err = m.Version()
	util.logger.Infow("current db version", "version", version)
	return version, nil
}

func (util MigrateUtil) migrateGoTo(location string, goToVersion uint) (version uint, err error) {
	m, err := migrate.New(location, util.config.DatabaseUrl)
	if err != nil {
		util.logger.Errorw("error in connection db", "err", err)
		return 0, err
	}
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		util.logger.Errorw("error in determining current migration will not apply any further migration ", "err", err)
		return 0, err
	} else if err == migrate.ErrNilVersion {
		util.logger.Infow("applying changes on fresh db", "targetVersion", goToVersion)
	} else if dirty {
		util.logger.Errorw("dirty db state found will not apply changes. please clean db manually first", "CurrentVersion", version)
		return 0, err
	} else {
		util.logger.Infow("applying changes over version ", "version", version, "targetVersion", goToVersion)
	}
	if err := m.Migrate(goToVersion); err != nil {
		util.logger.Errorw("error in applying migration", "err", err)
		return 0, err
	}
	version, dirty, err = m.Version()
	util.logger.Infow("current db version", "version", version)
	return version, nil
}
