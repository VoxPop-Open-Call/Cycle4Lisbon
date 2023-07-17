package dex

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/server/dex/sqlconnector"
	"github.com/dexidp/dex/server"
	"github.com/dexidp/dex/storage"
	"github.com/dexidp/dex/storage/sql"
	"github.com/sirupsen/logrus"
)

// Opens a connection to a postgres database.
func OpenStorage(config *StorageConfig) (storage.Storage, error) {
	pg := &sql.Postgres{
		NetworkDB: sql.NetworkDB{
			Database: config.Database,
			User:     config.User,
			Password: config.Password,
			Host:     config.Host,
			Port:     config.Port,
		},
		SSL: sql.SSL{
			Mode: config.SSL,
		},
	}

	return pg.Open(nil)
}

type utcFormatter struct {
	f logrus.Formatter
}

func (f *utcFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return f.f.Format(e)
}

type DexConfig struct {
	Store      storage.Storage // Dex storage
	ConfigFile string          // Dex config file
	UsersDbDsn string          // DSN of the database where users are stored
}

// Initializes the Dex server, and returns the handler and config.
func New(conf *DexConfig) (http.Handler, *Config, error) {
	// Add our custom sql connector to the Dex connectors list.
	server.ConnectorsConfig["sql"] = func() server.ConnectorConfig {
		return &sqlconnector.Config{
			DSN: conf.UsersDbDsn,
		}
	}

	config, err := loadConfig(conf.ConfigFile)
	if err != nil {
		return nil, nil,
			fmt.Errorf("failed to read config file %s: %v", conf.ConfigFile, err)
	}

	var formatter utcFormatter
	formatter.f = &logrus.TextFormatter{DisableColors: true}
	logger := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: &formatter,
		Level:     logrus.InfoLevel,
	}

	store := conf.Store
	store = storage.WithStaticClients(store, config.StaticClients)

	storageConnectors := make([]storage.Connector, len(config.StaticConnectors))
	for i, staticConnector := range config.StaticConnectors {
		storageConnector, err := toStorageConnector(staticConnector)
		if err != nil {
			return nil, nil,
				fmt.Errorf("failed to initialize storage connectors: %v", err)
		}
		storageConnectors[i] = storageConnector
	}
	store = storage.WithStaticConnectors(store, storageConnectors)

	now := func() time.Time { return time.Now().UTC() }

	logger.Infof("dex issuer: %s", config.Issuer)

	serverConfig := server.Config{
		Issuer:                 config.Issuer,
		Storage:                store,
		Logger:                 logger,
		Now:                    now,
		SkipApprovalScreen:     true,
		SupportedResponseTypes: []string{"code", "token", "id_token"},
		AllowedOrigins:         []string{"*"},
		PasswordConnector:      "custom",
	}

	refreshTokenPolicy, err := server.NewRefreshTokenPolicy(
		logger, false, "", "", "",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid refresh token policy: %v", err)
	}
	serverConfig.RefreshTokenPolicy = refreshTokenPolicy

	serv, err := server.NewServer(context.Background(), serverConfig)

	return serv, config, err
}
