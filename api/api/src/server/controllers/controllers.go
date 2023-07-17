package controllers

import (
	"errors"
	"regexp"

	"bitbucket.org/pensarmais/cycleforlisbon/src/aws"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/jobs"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/access"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/middleware"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/gobutil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/latlon"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/stringutil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var orderByValidationRegex = regexp.MustCompile(
	`^(\w+, ?)*\w+ (asc|desc)$`,
)

// customValidations to register with gin's validation engine.
var customValidations = map[string]validator.Func{
	"order_by_clause": func(fl validator.FieldLevel) bool {
		return orderByValidationRegex.MatchString(fl.Field().String())
	},
}

type Store struct {
	Users           *UserController
	Password        *PasswordController
	Initiatives     *InitiativeController
	SDGs            *SDGController
	Institutions    *InstitutionController
	Trips           *TripController
	Achievements    *AchievementController
	POIs            *POIController
	Leaderboard     *LeaderboardController
	ExternalContent *ExternalContentController
	FCMTokens       *FCMTokenController
	Languages       *LanguageController
	Metrics         *MetricsController
}

func NewStore(
	db *gorm.DB,
	acl *access.ACL,
	wrkr *worker.Worker,
	aws *aws.Client,
	geocoder *latlon.Geocoder,
	serverBaseURL string,
) *Store {
	// Register custom validation tags
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		for k, f := range customValidations {
			v.RegisterValidation(k, f)
		}
	}

	users := &UserController{db, acl, serverBaseURL, aws.S3}
	registerAllRules(users, acl)

	password := &PasswordController{db, aws.SES}

	initiatives := &InitiativeController{db, acl, aws.S3}
	registerAllRules(initiatives, acl)

	sdgs := &SDGController{db, serverBaseURL}

	institutions := &InstitutionController{db, acl, aws.S3}
	registerAllRules(institutions, acl)

	trips := &TripController{
		db, acl, wrkr, geocoder,
		gobutil.NewGobCodec[jobs.UpdateAchievementsArgs](),
	}
	registerAllRules(trips, acl)

	achievements := &AchievementController{db, serverBaseURL}

	pois := &POIController{db, acl}
	registerAllRules(pois, acl)

	leaderboard := &LeaderboardController{db}

	external := &ExternalContentController{db, acl}
	registerAllRules(external, acl)

	fcm := &FCMTokenController{db}

	languages := &LanguageController{db}

	metrics := &MetricsController{db, acl}
	registerAllRules(metrics, acl)

	return &Store{
		Users:           users,
		Password:        password,
		Initiatives:     initiatives,
		SDGs:            sdgs,
		Institutions:    institutions,
		Trips:           trips,
		Achievements:    achievements,
		POIs:            pois,
		Leaderboard:     leaderboard,
		ExternalContent: external,
		FCMTokens:       fcm,
		Languages:       languages,
		Metrics:         metrics,
	}
}

// authorizer interface is used to control access to resources by authorized
// entities.
// Should return true iff ent has permission to perform action on res.
type authorizer interface {
	Authorize(ent any, action string, res any) bool
}

type rule struct {
	ent, res any
	action   string
	f        access.AuthFunc
}

func registerAllRules(r interface {
	Rules() []rule
}, acl *access.ACL) {
	for _, rule := range r.Rules() {
		acl.Register(rule.ent, rule.action, rule.res, rule.f)
	}
}

// scheduler interface contains the workers schedule method.
//
// This interface allows mocking the worker during testing.
type scheduler interface {
	Schedule(*worker.TaskConfig) error
}

type Pagination struct {
	// Limit is the maximum number of records to be returned.
	// The API doesn't enforce a limit, it's a responsibility of the client.
	Limit int `form:"limit"`

	// Offset is the number of records to skip when retrieving.
	Offset int `form:"offset"`
}

type orderBy string

// ToSnakeCase converts the orderBy string from camelCase (used by the API) to
// snake_case (which is employed by gorm for the column names).
func (o orderBy) ToSnakeCase() string {
	return stringutil.CamelToSnake(string(o))
}

type Sort struct {
	// OrderBy specifies the sorting order of the records returned by List
	// methods.
	OrderBy orderBy `form:"orderBy,default=id asc" binding:"omitempty,order_by_clause" example:"foo,bar asc" default:"id asc"`
}

// PresignedResponse contains a presigned URL to perform an action on an object
// in an S3 bucket, and the expected HTTP method.
type PresignedResponse struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}

// tokenClaims retrieves the token claims from the gin context.
func tokenClaims(ctx *gin.Context) (*middleware.Claims, error) {
	val, ok := ctx.Get(middleware.TokenClaimsKey)
	if !ok {
		return nil, errors.New("token claims not present in context")
	}

	claims, ok := val.(middleware.Claims)
	if !ok {
		return nil, errors.New("failed to assert type of token claims")
	}

	return &claims, nil
}

// tokenUser retrieves the user identified by the token claims in the context.
func tokenUser(ctx *gin.Context, db *gorm.DB) (models.User, error) {
	claims, err := tokenClaims(ctx)
	if err != nil {
		return models.User{}, err
	}

	user, err := query.Users.FromClaims(claims, db)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, httputil.NewErrorMsg(
			httputil.TokenUserNotFound,
			"the token is valid, but the subject doesn't exist",
		)
	}

	return user, err
}

func resourceNotFoundErr(name string) httputil.Error {
	return httputil.NewErrorMsg(
		httputil.RecordNotFound,
		name+" not found",
	)
}
