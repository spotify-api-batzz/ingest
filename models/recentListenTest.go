package models

import (
	"spotify/utils"
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

func (r *RecentListen) TableName() string {
	return "recent_listens"
}

func (r *RecentListen) TableColumns() []string {
	return []string{"id", "song_id", "user_id", "played_at", "created_at", "updated_at"}
}

func (r *RecentListen) ToSlice() []interface{} {
	slice := make([]interface{}, 6)
	slice[0] = r.ID
	slice[1] = r.SongID
	slice[2] = r.UserID
	slice[3] = r.PlayedAt.Format(time.RFC3339)
	slice[4] = utils.Now()
	slice[5] = utils.Now()

	return slice
}
