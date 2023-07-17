package server

import (
	"fmt"
	"net/http"

	"bitbucket.org/pensarmais/cycleforlisbon/docs"
	"bitbucket.org/pensarmais/cycleforlisbon/src/aws"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/controllers"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/dex"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/middleware"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/route"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/latlon"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"github.com/dexidp/dex/storage"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

//	@title					Cycle for Lisbon API
//	@version				0.1
//	@licence.name			None
//	@BasePath				/api
//	@description.markdown	api

//	@securitydefinitions.apikey	AuthHeader
//	@description				Set the authorization header: `bearer` followed by the oidc token.
//	@in							header
//	@name						authorization

//	@securitydefinitions.oauth2.accessCode	OIDCToken
//	@description							OAuth 2.0 authentication flow. **Note:** only works on localhost.
//	@authorizationUrl						http://localhost:8080/dex/auth
//	@tokenUrl								http://localhost:8080/dex/token
//	@scope.openid
//	@scope.profile
//	@scope.email
//	@scope.offline_access

type Config struct {
	ApiHost       string
	ServerBaseURL string
	DexConfigFile string
	DB            *gorm.DB
	DbDsn         string
	DexStorage    storage.Storage
	Worker        *worker.Worker
	AWS           *aws.Client
	Geocoder      *latlon.Geocoder
}

// New initializes both Dex and Api handlers in a multiplexer.
func New(config *Config) (*http.Server, error) {
	dexHandler, dexConfig, err := initDex(
		config.DexStorage,
		config.DexConfigFile,
		config.DbDsn,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init Dexidp: %v", err)
	}

	apiHandler := initApi(
		config,
		dexConfig.Issuer,
		clientIds(dexConfig.StaticClients),
	)

	srvMux := http.NewServeMux()

	srvMux.Handle("/api/", apiHandler)
	srvMux.Handle("/docs/", apiHandler)
	srvMux.Handle("/public/", apiHandler)

	srvMux.Handle("/dex", dexHandler)
	srvMux.Handle("/dex/", dexHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: srvMux,
	}

	return srv, err
}

// initDex initializes the Dex identity provider and returns the handler.
func initDex(
	dexStore storage.Storage,
	dexConfigFile string,
	dbDsn string,
) (http.Handler, *dex.Config, error) {
	dexHandler, config, err := dex.New(&dex.DexConfig{
		Store:      dexStore,
		ConfigFile: dexConfigFile,
		UsersDbDsn: dbDsn,
	})

	return &httputil.CorsHandler{Handler: dexHandler}, config, err
}

// initApi initializes the api router using `gin` and returns the handler.
func initApi(
	config *Config,
	dexIssuerUrl string,
	dexClientIds []string,
) http.Handler {
	router := gin.New()
	router.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/api/health"),
		gin.Recovery(),
	)

	router.Use(middleware.Sentry())
	router.Use(middleware.Error())
	router.Use(middleware.CORS())

	auth, _ := middleware.Auth(dexIssuerUrl, dexClientIds)

	acl := access.New()
	store := controllers.NewStore(
		config.DB,
		acl,
		config.Worker,
		config.AWS,
		config.Geocoder,
		config.ServerBaseURL,
	)

	api := router.Group("/api")
	{
		api.HEAD("/health", healthCheck(config.DB))
		api.GET("/health", healthCheck(config.DB))

		route.Users(api, auth, store)
		route.Password(api, auth, store)
		route.Initiatives(api, auth, store)
		route.SDGs(api, auth, store)
		route.Institutions(api, auth, store)
		route.Trips(api, auth, store)
		route.Achievements(api, auth, store)
		route.POIs(api, auth, store)
		route.Leaderboard(api, auth, store)
		route.ExternalContent(api, auth, store)
		route.FCM(api, auth, store)
		route.Languages(api, store)
		route.Metrics(api, auth, store)
	}

	// Serve docs
	docs.SwaggerInfo.Host = config.ApiHost
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve static files
	router.Static("/public", "./public")

	return router
}

// healthCheck
//
//	@Summary	Perform a health check on this API
//	@Produce	json
//	@Success	200
//	@Failure	500	{object}	middleware.ApiError
//	@Router		/health [get]
//	@Router		/health [head]
func healthCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Raw("select 1").Error; err != nil {
			c.Error(httputil.NewError(
				httputil.InternalServerError,
				err,
			))
			return
		}

		c.Status(http.StatusOK)
	}
}

func clientIds(clients []storage.Client) []string {
	result := make([]string, len(clients))
	for i, client := range clients {
		result[i] = client.ID
	}
	return result
}
