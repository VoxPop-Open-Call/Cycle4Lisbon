package controllers

import (
	"database/sql"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LeaderboardController struct {
	db *gorm.DB
}

type LeaderboardResult struct {
	Entries      []query.LeaderboardEntry `json:"entries"`
	UserPosition int                      `json:"userPosition"`
}

// List users in the leaderboard.
//
//	@Summary	List top 10 users in the leaderboard
//	@Tags		leaderboard
//	@Produce	json
//	@Security	OIDCToken
//	@Security	AuthHeader
//	@Success	200			{object}	LeaderboardResult
//	@Failure	400,401,500	{object}	middleware.ApiError
//	@Router		/leaderboard  [get]
func (c *LeaderboardController) List(ctx *gin.Context) (LeaderboardResult, error) {
	user, err := tokenUser(ctx, c.db)
	if err != nil {
		return LeaderboardResult{}, err
	}

	res := LeaderboardResult{}

	err = c.db.Transaction(func(tx *gorm.DB) error {
		top, err := query.Leaderboard.Top(tx)
		if err != nil {
			return err
		}
		res.Entries = top

		userPosition, err := query.Leaderboard.PositionOf(user.ID.String(), tx)
		res.UserPosition = userPosition
		return err
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})

	return res, err
}
