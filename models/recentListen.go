package models

import (
	"spotify/utils"
	"time"
)

type RecentListen struct {
	ID        string     `db:"id"`
	SongID    string     `db:"song_id"`
	UserID    string     `db:"user_id"`
	PlayedAt  utils.Time `db:"played_at"`
	CreatedAt utils.Time `db:"created_at"`
	UpdatedAt utils.Time `db:"updated_at"`

	Song
	User
}

func NewRecentListen(songID string, userID string, playedAt time.Time) RecentListen {
	return RecentListen{
		ID:        utils.GenerateUUID(),
		SongID:    songID,
		UserID:    userID,
		PlayedAt:  utils.Time{playedAt},
		CreatedAt: utils.NewTime(),
		UpdatedAt: utils.NewTime(),
	}
}

func (r *RecentListen) TableName() string {
	return "recent_listens"
}
