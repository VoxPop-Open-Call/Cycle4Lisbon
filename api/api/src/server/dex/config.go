package dex

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/dexidp/dex/server"
	"github.com/dexidp/dex/storage"
)

type StorageConfig struct {
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     uint16 `json:"port"`
	SSL      string `json:"ssl"`
}

type Config struct {
	Issuer           string           `json:"issuer"`
	StaticClients    []storage.Client `json:"staticClients"`
	StaticConnectors []connector      `json:"connectors"`
}

func loadConfig(configFile string) (*Config, error) {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", configFile, err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file %s: %v", configFile, err)
	}

	return &config, nil
}

type connector struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config server.ConnectorConfig `json:"config"`
}

// UnmarshalJSON allows Connector to implement the unmarshaler interface to
// dynamically determine the type of the connector config.
func (c *connector) UnmarshalJSON(b []byte) error {
	var conn struct {
		Type string `json:"type"`
		Name string `json:"name"`
		ID   string `json:"id"`

		Config json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(b, &conn); err != nil {
		return fmt.Errorf("parse connector: %v", err)
	}
	f, ok := server.ConnectorsConfig[conn.Type]
	if !ok {
		return fmt.Errorf("unknown connector type %q", conn.Type)
	}

	connConfig := f()
	if len(conn.Config) != 0 {
		data := []byte(conn.Config)
		if isExpandEnvEnabled() {
			// Caution, we're expanding in the raw JSON/YAML source. This may not be what the admin expects.
			data = []byte(os.ExpandEnv(string(conn.Config)))
		}
		if err := json.Unmarshal(data, connConfig); err != nil {
			return fmt.Errorf("parse connector config: %v", err)
		}
	}
	*c = connector{
		Type:   conn.Type,
		Name:   conn.Name,
		ID:     conn.ID,
		Config: connConfig,
	}
	return nil
}

func toStorageConnector(c connector) (storage.Connector, error) {
	data, err := json.Marshal(c.Config)
	if err != nil {
		return storage.Connector{},
			fmt.Errorf("failed to marshal connector config: %v", err)
	}

	return storage.Connector{
		ID:     c.ID,
		Type:   c.Type,
		Name:   c.Name,
		Config: data,
	}, nil
}

// isExpandEnvEnabled returns if os.ExpandEnv should be used for each storage and connector config.
// Disabling this feature avoids surprises e.g. if the LDAP bind password contains a dollar character.
// Returns false if the env variable "DEX_EXPAND_ENV" is a falsy string, e.g. "false".
// Returns true if the env variable is unset or a truthy string, e.g. "true", or can't be parsed as bool.
func isExpandEnvEnabled() bool {
	enabled, err := strconv.ParseBool(os.Getenv("DEX_EXPAND_ENV"))
	if err != nil {
		// Unset, empty string or can't be parsed as bool: Default = true.
		return true
	}
	return enabled
}
