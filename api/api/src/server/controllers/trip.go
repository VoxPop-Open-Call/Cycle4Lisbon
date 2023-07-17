package controllers

import (
	"crypto/sha256"
	"errors"
	"log"
	"math"
	"regexp"

	"bitbucket.org/pensarmais/cycleforlisbon/src/achievements"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/jobs"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/gobutil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/gpx"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/httputil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/latlon"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TripController struct {
	db       *gorm.DB
	acl      authorizer
	tasks    scheduler
	geocoder interface {
		ReverseAddr(coords latlon.Coords) string
	}
	jobCodec *gobutil.GobCodec[jobs.UpdateAchievementsArgs]
}

func (TripController) Rules() []rule {
	return []rule{
		{models.User{}, models.Trip{}, "get", func(ent, res any) bool {
			user := ent.(models.User)
			trip := res.(models.Trip)
			return user.Admin || user.ID == trip.UserID
		}},
	}
}

var duplicateGPXRegex = regexp.MustCompile(
	"duplicate key value violates unique constraint \"trips_gpx_hash_key\"",
)

type ListTripsFilters struct {
	Pagination
	Sort
	// TimeFrom filters trips uploaded after this time.
	TimeFrom string `form:"timeFrom" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00" example:"2023-03-30T17:23:57+02:00"`
	// TimeTo filters trips uploaded before this time.
	TimeTo string `form:"timeTo" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00" example:"2023-03-30T17:23:57+02:00"`
}

// List all trips of a user.
//
//	@Summary		List all trips
//	@Description	Admin users have access to all trips, and regular users only to their own.
//	@Tags			trips
//	@Produce		json
//	@Security		OIDCToken
//	@Security		AuthHeader
//	@Param			filters		query		ListTripsFilters	false	"Filters"
//	@Success		200			{array}		models.Trip
//	@Failure		400,401,500	{object}	middleware.ApiError
//	@Router			/trips  [get]
func (c *TripController) List(
	filters ListTripsFilters,
	ctx *gin.Context,
) ([]models.Trip, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return nil, err
	}

	tx := c.db

	if !user.Admin {
		tx = tx.Where("user_id = ?", user.ID)
	} else {
		tx = tx.Joins("User")
	}

	if filters.TimeFrom != "" {
		tx = tx.Where("created_at >= ?", filters.TimeFrom)
	}
	if filters.TimeTo != "" {
		tx = tx.Where("created_at <= ?", filters.TimeTo)
	}

	var trips []models.Trip
	err = tx.
		Limit(filters.Limit).
		Offset(filters.Offset).
		Order(filters.OrderBy.ToSnakeCase()).
		Find(&trips).Error

	return trips, err
}

// Get retrieves a trip.
//
//	@Summary	Retrieve a trip by Id
//	@Tags		trips
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id				path		string	true	"Trip Id"	Format(UUID)
//	@Success	200				{object}	models.Trip
//	@Failure	400,401,404,500	{object}	middleware.ApiError
//	@Router		/trips/{id} [get]
func (c *TripController) Get(id string, ctx *gin.Context) (models.Trip, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return models.Trip{}, err
	}

	tx := c.db
	if user.Admin {
		tx = tx.Joins("User")
	}

	var trip models.Trip
	if err = tx.Model(&models.Trip{}).
		Joins("Initiative").
		First(&trip, "trips.id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Trip{}, resourceNotFoundErr("trip")
		}

		return models.Trip{}, err
	}

	if ok := c.acl.Authorize(
		user, "get", trip,
	); !ok {
		return models.Trip{}, httputil.NewErrorMsg(
			httputil.Forbidden,
			httputil.ForbiddenMessage,
		)
	}

	return trip, nil
}

// Download retrieves a trip's GPX file.
//
//	@Summary	Download a trip's GPX file by Id
//	@Tags		trips
//	@Produce	json
//	@Produce	application/gpx+xml
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		id				path		string	true	"Trip Id"	Format(UUID)
//	@Success	200				{file}		"The binary data of the GPX file"
//	@Failure	400,401,404,500	{object}	middleware.ApiError
//	@Router		/trips/{id}/file [get]
func (c *TripController) Download(id string, ctx *gin.Context) ([]byte, string, error) {
	trip, err := c.Get(id, ctx)
	ctx.Header("Content-Disposition", "attachment; filename="+id+".gpx")
	return trip.GPX, "application/gpx+xml", err
}

func isValid(trip *gpx.GPX) (valid bool, reason string) {
	if len(trip.Track.Segment) < 2 {
		return false, "the track must have at least 2 points"
	}

	// TODO: roughly validate if trip was done in a bicycle.
	// - average speed in motion?
	// - take inclination into account? average speed climbing vs descending?
	return true, ""
}

// addAddresses fetches the addresses of the start and end point from their
// coordinates in the gpx file, and adds them to the trip.
func (c *TripController) addAddresses(trip *models.Trip, gpxTrip *gpx.GPX) {
	start := gpxTrip.StartPoint()
	end := gpxTrip.EndPoint()
	trip.StartLat = start.Lat
	trip.StartLon = start.Lon
	trip.EndLat = end.Lat
	trip.EndLon = end.Lon

	trip.StartAddr = c.geocoder.ReverseAddr(
		latlon.Coords{Lat: start.Lat, Lon: start.Lon},
	)
	trip.EndAddr = c.geocoder.ReverseAddr(
		latlon.Coords{Lat: end.Lat, Lon: end.Lon},
	)
}

// updateStats credits the user's current initiative and updates the user's
// stats.
func updateStats(trip *models.Trip, user *models.User, tx *gorm.DB) error {
	ratio, err := query.Settings.KilometersCreditsRatio(tx)
	if err != nil {
		return err
	}

	trip.Credits = math.Floor(trip.Distance / float64(ratio))
	if err = tx.Save(&trip).Error; err != nil {
		return err
	}

	if trip.InitiativeID != nil {
		if err = query.Initiatives.
			Credit(trip.InitiativeID.String(), trip.Credits, tx); err != nil {
			// Don't return an error if the initiative has ended.
			if errors.Is(err, query.ErrInitiativeEnded) {
				log.Println("not crediting selected initiative because it has ended")
			} else {
				return err
			}
		}
	}

	return query.Users.UpdateStats(user, trip.Distance, trip.Credits, tx)
}

func (c *TripController) scheduleAchievmentsUpdate(user models.User, tx *gorm.DB) error {
	initiatives, err := query.Users.InitiativeCount(user.ID.String(), tx)
	if err != nil {
		return err
	}

	args, err := c.jobCodec.Encode(jobs.UpdateAchievementsArgs{
		UserID: user.ID,
		State: achievements.State{
			Rides:       user.TripCount,
			Distance:    user.TotalDist,
			Credits:     user.Credits,
			Initiatives: initiatives,
		},
	})
	if err != nil {
		return err
	}

	return c.tasks.Schedule(&worker.TaskConfig{
		JobName: jobs.UpdateAchievements,
		Args:    args,
	})
}

// Upload trip (gpx file).
//
//	@Summary	Upload trip (gpx file)
//	@Tags		trips
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Param		file		formData	file	true	"Params"
//	@Success	200			{object}	models.Trip
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/trips [post]
func (c *TripController) Upload(
	data []byte,
	ctx *gin.Context,
) (models.Trip, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return models.Trip{}, err
	}

	gpxTrip := new(gpx.GPX)
	if err = gpxTrip.Unmarshal(data); err != nil {
		return models.Trip{},
			httputil.NewError(httputil.InvalidGPXFile, err)
	}

	hash := sha256.New()
	if _, err = hash.Write(data); err != nil {
		return models.Trip{}, err
	}

	duration, durationInMotion := gpxTrip.Duration()
	valid, reason := isValid(gpxTrip)
	trip := &models.Trip{
		GPX:              data,
		GPXHash:          hash.Sum(nil),
		IsValid:          valid,
		NotValidReason:   reason,
		Distance:         gpxTrip.Distance(),
		Duration:         duration.Seconds(),
		DurationInMotion: durationInMotion.Seconds(),
		UserID:           user.ID,
		InitiativeID:     user.InitiativeID,
	}

	if trip.IsValid {
		c.addAddresses(trip, gpxTrip)
	}

	err = c.db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(trip).Error; err != nil {
			if duplicateGPXRegex.MatchString(err.Error()) {
				return httputil.NewErrorMsg(
					httputil.DuplicatedGPXFile,
					"The provided GPX file has already been uploaded",
				)
			}
			return err
		}

		if !trip.IsValid {
			// Don't credit invalid trips, but don't error either.
			log.Printf("not crediting trip because it failed validation: %s",
				trip.NotValidReason)
			return nil
		}

		if err = updateStats(trip, &user, tx); err != nil {
			return err
		}

		return c.scheduleAchievmentsUpdate(user, tx)
	})

	return *trip, err
}
