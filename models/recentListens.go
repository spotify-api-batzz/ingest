package models

import (
	"time"

	"github.com/google/uuid"
)

type RecentListen struct {
	ID        uuid.UUID `db:"id"`
	SongID    string    `db:"song_id"`
	UserID    string    `db:"user_id"`
	PlayedAt  time.Time `db:"played_at"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Song
	User
}

func NewRecentListen(songID string, userID string, playedAt time.Time) RecentListen {
	return RecentListen{
		ID:       uuid.New(),
		SongID:   songID,
		UserID:   userID,
		PlayedAt: playedAt,
	}
}

func (r *RecentListen) ToSlice() []interface{} {
	slice := make([]interface{}, 6)
	slice[0] = r.ID
	slice[1] = r.SongID
	slice[2] = r.UserID
	slice[3] = r.PlayedAt
	slice[4] = time.Now().UTC().Format(time.RFC3339)
	slice[5] = time.Now().UTC().Format(time.RFC3339)

	return slice
}
