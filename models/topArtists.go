package models

import (
	"spotify/utils"
)

type TopArtist struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`

	CreatedAt utils.Time `db:"created_at"`
	UpdatedAt utils.Time `db:"updated_at"`
}

func NewTopArtist(userID string) TopArtist {
	return TopArtist{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		CreatedAt: utils.NewTime(),
		UpdatedAt: utils.NewTime(),
	}
}

func (r *TopArtist) TableName() string {
	return "top_artists"
}
