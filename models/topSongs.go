package models

import (
	"spotify/utils"
)

type TopSong struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	CreatedAt utils.Time `db:"created_at"`
	UpdatedAt utils.Time `db:"updated_at"`
}

func (r *TopSong) TableName() string {
	return "top_songs"
}

func NewTopSong(userID string) TopSong {
	return TopSong{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		CreatedAt: utils.NewTime(),
		UpdatedAt: utils.NewTime(),
	}
}
