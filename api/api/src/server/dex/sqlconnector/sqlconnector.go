/*
"Heavily inspired" by this pull request, without some features we don't need:

https://github.com/dexidp/dex/pull/1833

Changes relative to that PR are:
- Remove some configuration options, mainly the option to use other
connectors besides postgres.
- Remove support for user groups.
- Revert to using bcrypt instead of passlib.
- Add support for uuid type id colunm.
- Remove SQL queries and column names from the config, and hardcode them here.
*/
package sqlconnector

import (
	"context"
	"encoding/json"
	"fmt"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/password"
	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/pkg/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// The SQL connnector finds the user based on the username and password given
// to it.
//
// There are two versions of the query, one keyed on username and one keyed on
// user ID. The latter is used to implement refresh.
//
// The refresh logic assumes that user IDs (whatever they may be) are immutable
// and never reused.

type Config struct {
	// The database DSN.
	DSN string

	// UsernamePrompt allows users to override the username attribute
	// (displayed in the username/password prompt). If unset, the handler will
	// use "Username".
	UsernamePrompt string `json:"usernamePrompt,omitempty"`
}

var userQuery = struct {
	// The actual SQL.
	QueryByName string
	QueryByID   string

	// The names of various columns.
	IDColumn       string
	EmailColumn    string
	NameColumn     string
	PasswordColumn string
}{
	QueryByName:    "select id, email, name, hashed_password from users where email=:username",
	QueryByID:      "select id, email, name, hashed_password from users where id=:userid",
	IDColumn:       "id",
	EmailColumn:    "email",
	NameColumn:     "id",
	PasswordColumn: "hashed_password",
}

func (c *Config) Open(id string, logger log.Logger) (connector.Connector, error) {
	conn, err := c.OpenConnector(logger)
	if err != nil {
		return nil, err
	}
	return connector.Connector(conn), nil
}

func (c *Config) OpenConnector(logger log.Logger) (interface {
	connector.Connector
	connector.PasswordConnector
	connector.RefreshConnector
}, error) {
	return c.openConnector(logger)
}

func (c *Config) openConnector(logger log.Logger) (*sqlConnector, error) {
	db, err := sqlx.Open("postgres", c.DSN)
	if err != nil {
		logger.Errorf("sql: cannot connect to %q with driver %q: %s",
			c.DSN, "postgres", err)
		return nil, err
	}
	return &sqlConnector{*c, db, logger}, nil
}

type sqlConnector struct {
	Config

	db *sqlx.DB

	logger log.Logger
}

type sqlRefreshData struct {
	UserID string `json:"userid"`
}

var (
	_ connector.PasswordConnector = (*sqlConnector)(nil)
	_ connector.RefreshConnector  = (*sqlConnector)(nil)
)

func (c *sqlConnector) identityFromRow(row map[string]interface{}) (ident connector.Identity, err error) {
	var ok bool

	id := row[userQuery.IDColumn]
	if idint, ok := id.(int); ok {
		ident.UserID = fmt.Sprintf("%d", idint)
	} else if idstr, ok := id.(string); ok {
		ident.UserID = idstr
	} else if iduuid, ok := id.([]uint8); ok {
		// uuid column is parsed as []uint8
		ident.UserID = string(iduuid)
	} else {
		fmt.Printf("id %s %T", id, id)

		err = fmt.Errorf("IDColumn %s must be a string, int or uuid",
			userQuery.IDColumn)
		return connector.Identity{}, err
	}

	if userQuery.EmailColumn != "" {
		ident.Email, ok = row[userQuery.EmailColumn].(string)
		if !ok {
			err = fmt.Errorf("sql: EmailColumn %s must be a string",
				userQuery.EmailColumn)
			return connector.Identity{}, err
		}
	}

	if userQuery.NameColumn != "" {
		name := row[userQuery.NameColumn]
		if namestr, ok := name.(string); ok {
			ident.Username = namestr
		} else if nameuuid, ok := id.([]uint8); ok {
			ident.Username = string(nameuuid)
		} else {
			err = fmt.Errorf("sql: NameColumn %s must be a string or uuid",
				userQuery.NameColumn)
			return connector.Identity{}, err
		}
	}

	return ident, nil
}

func (c *sqlConnector) Login(ctx context.Context, s connector.Scopes,
	username, passwd string) (ident connector.Identity,
	validPass bool, err error) {

	rows, err := c.db.NamedQueryContext(ctx, userQuery.QueryByName,
		map[string]interface{}{
			"username": username,
		})
	if err != nil {
		return connector.Identity{}, false, err
	}
	defer rows.Close()

	if !rows.Next() {
		err = rows.Err()
		return connector.Identity{}, false, err
	}

	row := map[string]interface{}{}

	err = rows.MapScan(row)
	if err != nil {
		return connector.Identity{}, false, err
	}

	ident, err = c.identityFromRow(row)
	if err != nil {
		return connector.Identity{}, false, err
	}

	passwdHash, ok := row[userQuery.PasswordColumn].(string)
	if !ok {
		err = fmt.Errorf("sql: PasswordColumn %s must be a string",
			userQuery.PasswordColumn)
		return connector.Identity{}, false, err
	}

	if !password.Check(passwd, passwdHash) {
		c.logger.Warnf("sql: incorrect password for user %s", ident.UserID)
		return connector.Identity{}, false, nil
	}

	if s.OfflineAccess {
		refresh := sqlRefreshData{
			UserID: ident.UserID,
		}

		if ident.ConnectorData, err = json.Marshal(refresh); err != nil {
			return connector.Identity{}, false,
				fmt.Errorf("sql: failed to marshal refresh data: %v", err)
		}
	}

	return ident, true, nil
}

func (c *sqlConnector) Refresh(ctx context.Context, s connector.Scopes,
	ident connector.Identity) (newIdent connector.Identity, err error) {

	var refreshData sqlRefreshData
	if err := json.Unmarshal(ident.ConnectorData, &refreshData); err != nil {
		return ident,
			fmt.Errorf("sql: failed to unmarshal internal data: %v", err)
	}

	rows, err := c.db.NamedQueryContext(ctx, userQuery.QueryByID,
		map[string]interface{}{
			"userid": refreshData.UserID,
		})
	if err != nil {
		return connector.Identity{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err == nil {
			err = fmt.Errorf("sql: user %q not found during refresh",
				refreshData.UserID)
		} else {
			err = rows.Err()
		}

		return connector.Identity{}, err
	}

	row := map[string]interface{}{}

	err = rows.MapScan(row)
	if err != nil {
		return connector.Identity{}, err
	}

	newIdent, err = c.identityFromRow(row)
	if err != nil {
		return connector.Identity{}, err
	}
	newIdent.ConnectorData = ident.ConnectorData

	return newIdent, nil
}

func (c *sqlConnector) Prompt() string {
	return c.UsernamePrompt
}
