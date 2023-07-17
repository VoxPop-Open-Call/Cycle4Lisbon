package query

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"gorm.io/gorm"
)

type metrics struct{}

var Metrics metrics

type PlatformMetrics struct {
	TotalInitiatives     int64   `json:"totalInitiatives"`
	CompletedInitiatives int64   `json:"completedInitiatives"`
	OngoingInitiatives   int64   `json:"ongoingInitiatives"`
	TotalCredits         float64 `json:"totalCledits"`
}

func (metrics) Platform(tx *gorm.DB) (PlatformMetrics, error) {
	metrics := PlatformMetrics{}

	tx.Find(&models.Initiative{}).
		Count(&metrics.TotalInitiatives)

	tx.Find(&models.Initiative{}).
		Where("enabled = true").
		Where(tx.
			Where("end_date < now()").
			Or("credits >= goal"),
		).
		Count(&metrics.CompletedInitiatives)

	tx.Find(&models.Initiative{}).
		Where("enabled = true").
		Where("end_date >= now()").
		Where("credits < goal").
		Count(&metrics.OngoingInitiatives)

	tx.Model(&models.User{}).
		Select("sum(credits)").
		Scan(&metrics.TotalCredits)

	return metrics, tx.Error
}

type ageGroups struct {
	AgeLt18   int64 `json:"age<18"`
	Age18To25 int64 `json:"18<=age<25"`
	Age25To30 int64 `json:"25<=age<30"`
	Age30To40 int64 `json:"30<=age<40"`
	Age40To60 int64 `json:"40<=age<60"`
	Age60To75 int64 `json:"60<=age<75"`
	AgeGte75  int64 `json:"age>=75"`
}

type genderCount struct {
	M int64 `json:"m"`
	F int64 `json:"f"`
	X int64 `json:"x"`
}

type UserMetrics struct {
	Total      int64   `json:"total"`
	AverageAge float64 `json:"aveAge"`

	// AgeGroups maps the number of users in each age range.
	AgeGroups ageGroups `json:"ageGroups"`
	// GenderCount maps the number of users of each gender.
	GenderCount genderCount `json:"genderCount"`
}

func (metrics) Users(tx *gorm.DB) (UserMetrics, error) {
	metrics := UserMetrics{}

	tx.Model(&models.User{}).Count(&metrics.Total)

	tx.Raw(`
		SELECT avg(age) FROM (
			SELECT date_part('year', age(birthday)) AS age
			FROM users
		) as ages
    `).Scan(&metrics.AverageAge)

	tx.Raw(`
		WITH user_age AS (
			SELECT date_part('year', age(birthday)) AS age FROM users
		)
		SELECT
			count(*) filter(WHERE age<18) AS "age_lt18",
			count(*) filter(WHERE age>=18 AND age<25) AS "age18_to25",
			count(*) filter(WHERE age>=25 AND age<30) AS "age25_to30",
			count(*) filter(WHERE age>=30 AND age<40) AS "age30_to40",
			count(*) filter(WHERE age>=40 AND age<60) AS "age40_to60",
			count(*) filter(WHERE age>=60 AND age<75) AS "age60_to75",
			count(*) filter(WHERE age>=75) AS "age_gte75"
		FROM user_age
	`).Scan(&metrics.AgeGroups)

	var genders []struct {
		Gender string
		Count  int64
	}
	tx.Raw(`
		SELECT gender, count(gender)
		FROM users
		GROUP BY gender
		HAVING gender IS NOT null
	`).Scan(&genders)
	metrics.GenderCount = toGenderCount(genders)

	return metrics, tx.Error
}

func toGenderCount(raw []struct {
	Gender string
	Count  int64
}) genderCount {
	result := genderCount{}
	for _, g := range raw {
		switch g.Gender {
		case "M":
			result.M = g.Count
		case "F":
			result.F = g.Count
		case "X":
			result.X = g.Count
		}
	}
	return result
}

func toMap(raw []struct {
	Code  string
	Count int64
}) map[string]int64 {
	result := make(map[string]int64)
	for _, e := range raw {
		result[e.Code] = e.Count
	}
	return result
}

type TripMetrics struct {
	Total          int64   `json:"total"`
	AverageDist    float64 `json:"averageDist"`
	AverageCredits float64 `json:"averageCredits"`
}

func (metrics) Trips(tx *gorm.DB) (TripMetrics, error) {
	metrics := TripMetrics{}

	tx.Find(&models.Trip{}).
		Where("is_valid = true").
		Count(&metrics.Total)

	tx.Raw(`
		SELECT avg(distance) FROM trips
		WHERE is_valid = true
    `).Scan(&metrics.AverageDist)

	tx.Raw(`
		SELECT avg(credits) FROM trips
		WHERE is_valid = true
    `).Scan(&metrics.AverageCredits)

	return metrics, tx.Error
}
