package seeders

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/types"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"gorm.io/gorm"
)

type user struct{}

var User user

func (user) Seed(db *gorm.DB) {
	var initiatives []models.Initiative
	db.Limit(3).Find(&initiatives)

	var users = []models.User{
		{
			Profile: models.Profile{
				Name:     "Stan Marsh",
				Username: "stan",
				Birthday: types.DatePtr("1997-08-13"),
				Gender:   "M",
			},
			Email:          "stanley@tegrityfarms.com",
			Verified:       true,
			Subject:        random.String(10),
			HashedPassword: "$2a$10$D9GbC0NrCIf2/gtTTzX87eoZ3J6Z/gt4NxtyS.SbRD4UVR1l6Y60q",
			Initiative:     &initiatives[0],
			TripCount:      5,
			TotalDist:      200,
			Credits:        10,
		},
		{
			Profile: models.Profile{
				Name:     "Alice",
				Birthday: types.DatePtr("1969-01-20"),
				Gender:   "F",
			},
			Email:          "alice@example.com",
			Subject:        random.String(10),
			HashedPassword: "$2a$10$TtpOQuvqqcgcCflgzp3gKeCIU2kKKP7i95bWva0qwVnf1Ehv7NFVe",
			Initiative:     &initiatives[1],
			TripCount:      1,
			TotalDist:      5,
			Credits:        0.2,
		},
		{
			Profile: models.Profile{
				Name:     "Bob",
				Username: "bob123",
				Birthday: types.DatePtr("1950-12-29"),
				Gender:   "M",
			},
			Email:          "bob@builders.gov",
			Subject:        random.String(10),
			HashedPassword: "$2a$10$PCDN5RlNbLZ51R3NfQEAl.tDBGXmEXYhAqH5cn0u/BkgtHp8hikK2",
			Initiative:     &initiatives[1],
			TripCount:      10,
			TotalDist:      500,
			Credits:        25,
		},
		{
			Profile: models.Profile{
				Name:     "Carl Sagan",
				Birthday: types.DatePtr("1934-11-09"),
				Gender:   "M",
			},
			Subject: random.String(10),
			Email:   "test@example.com",
		},
		{
			// user returned by Dex's mock connector
			Email:    "kilgore@kilgore.trout",
			Verified: true,
			Subject:  "Cg0wLTM4NS0yODA4OS0wEgRtb2Nr",
			Profile: models.Profile{
				Name:     "Kilgore Trout",
				Birthday: types.DatePtr("1998-12-04"),
				Gender:   "M",
			},
			Initiative: &initiatives[2],
			TripCount:  50,
			TotalDist:  1500,
			Credits:    100,
		},
	}

	db.CreateInBatches(users, 50)
}
