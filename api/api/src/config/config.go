package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ENV string `default:"development"`

	SCHEME   string `default:"http"`
	API_HOST string `default:"localhost:8080"`

	API_ADMIN_EMAIL  string
	API_ADMIN_PASSWD string

	DB_HOST     string `default:"localhost"`
	DB_PORT     string `default:"5432"`
	DB_USER     string `default:"postgres"`
	DB_PASSWORD string `default:"password"`
	DB_NAME     string `default:"cycleforlisbon"`
	DB_SSL      string `default:"disable"`

	DEX_DB_HOST     string `default:"localhost"`
	DEX_DB_PORT     uint16 `default:"5432"`
	DEX_DB_USER     string `default:"postgres"`
	DEX_DB_PASSWORD string `default:"password"`
	DEX_DB_NAME     string `default:"dex"`
	DEX_DB_SSL      string `default:"disable"`

	AWS_REGION            string `default:"eu-west-1"`
	AWS_ACCESS_KEY_ID     string
	AWS_SECRET_ACCESS_KEY string
	AWS_S3_BUCKET         string
	AWS_SES_DOMAIN        string
	AWS_SES_FROM_NAME     string `default:"Cycle for Lisbon"`

	SENTRY_DSN string

	GOOGLE_API_KEY string

	GIN_MODE string `default:"debug"`
}

type configErr struct {
	errs []error
}

func (e *configErr) Error() string {
	sb := strings.Builder{}
	for _, err := range e.errs {
		sb.WriteString(err.Error())
		sb.WriteRune('\n')
	}
	return sb.String()
}

// get returns the first value to be defined in the following order: m[key], an
// environment variable named by the key, the defaultVal.
func get(m map[string]string, key string, defaultVal string) string {
	val, ok := m[key]
	if !ok {
		val, ok = os.LookupEnv(key)
		if !ok {
			val = defaultVal
		}
	}
	return val
}

// Load the config from the file in `path` and env variables.
// In case of conflict, the file takes precedence.
func Load(path string) (*Config, error) {
	config, _ := godotenv.Read(path)

	result := &Config{}
	fields := reflect.TypeOf(*result)
	n := fields.NumField()
	var errs []error
	for i := 0; i < n; i++ {
		key := fields.Field(i).Name
		defaultVal := fields.Field(i).Tag.Get("default")
		value := get(config, key, defaultVal)

		typeName := fields.Field(i).Type.Name()
		if typeName[:4] == "uint" {
			uintVal, err := strconv.ParseUint(value, 10, 16)
			if err != nil {
				errs = append(errs, err)
			}
			reflect.ValueOf(result).Elem().FieldByName(key).SetUint(uintVal)
		} else {
			reflect.ValueOf(result).Elem().FieldByName(key).SetString(value)
		}
	}

	if len(errs) > 0 {
		return nil, &configErr{errs}
	}
	return result, nil
}

func (c *Config) DbDsn() string {
	return "host=" + c.DB_HOST + " " +
		"port=" + c.DB_PORT + " " +
		"user=" + c.DB_USER + " " +
		"password=" + c.DB_PASSWORD + " " +
		"dbname=" + c.DB_NAME + " " +
		"sslmode=" + c.DB_SSL
}

func (c *Config) ServerBaseURL() string {
	return fmt.Sprintf("%s://%s", c.SCHEME, c.API_HOST)
}
