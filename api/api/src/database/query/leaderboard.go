package query

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type leaderboard struct{}

var Leaderboard leaderboard

type LeaderboardEntry struct {
	Position  int       `json:"position"`
	ID        uuid.UUID `json:"userId"`
	Name      string    `json:"name,omitempty"`
	Username  string    `json:"username,omitempty"`
	TripCount uint      `json:"tripCount"`
	TotalDist float64   `json:"totalDist"`
	Credits   float64   `json:"credits"`
}

func (leaderboard) Top(db *gorm.DB) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry
	err := db.Raw(`
		SELECT id, name, username, total_dist, trip_count, credits,
			row_number() over(ORDER BY total_dist DESC) as position
		FROM users
		ORDER BY total_dist DESC
		LIMIT 10
	`).Find(&entries).Error

	return entries, err
}

func (leaderboard) PositionOf(userID string, db *gorm.DB) (int, error) {
	var pos int
	err := db.Raw(`
		SELECT position
		FROM (
			SELECT id, 
				row_number() over(order by total_dist desc) as position
			FROM users
		) AS top
		WHERE id = ?
	`, userID).Scan(&pos).Error
	return pos, err
}
