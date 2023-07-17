package controllers

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MetricsController struct {
	db  *gorm.DB
	acl authorizer
}

func (MetricsController) Rules() []rule {
	return []rule{
		{models.User{}, Metrics{}, "get", func(ent, res any) bool {
			return ent.(models.User).Admin
		}},
	}
}

type Metrics struct {
	Platform query.PlatformMetrics `json:"platform"`
	Users    query.UserMetrics     `json:"users"`
	Trips    query.TripMetrics     `json:"trips"`
}

// Retrieve metrics.
//
//	@Summary	Retrieve metrics
//	@Tags		metrics
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Success	200					{object}	Metrics
//	@Failure	400,401,403,404,500	{object}	middleware.ApiError
//	@Router		/metrics [get]
func (c *MetricsController) Get(ctx *gin.Context) (Metrics, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return Metrics{}, err
	}

	if ok := c.acl.Authorize(
		user, "get", Metrics{},
	); !ok {
		return Metrics{}, httputil.NewErrorMsg(
			httputil.AdminAccessRequired,
			httputil.AdminRequiredMessage,
		)
	}

	metrics := Metrics{}
	err = c.db.Transaction(func(tx *gorm.DB) error {
		metrics.Platform, err = query.Metrics.Platform(c.db)
		if err != nil {
			return err
		}

		metrics.Users, err = query.Metrics.Users(c.db)
		if err != nil {
			return err
		}

		metrics.Trips, err = query.Metrics.Trips(c.db)
		return err
	})

	return metrics, err
}
