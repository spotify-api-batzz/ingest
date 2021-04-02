package models

import (
	"spotify/utils"
	"time"

	"github.com/google/uuid"
)

type TopSong struct {
	ID        uuid.UUID `db:"id"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *TopSong) TableName() string {
	return "top_songs"
}

func (r *TopSong) TableColumns() []string {
	return []string{"id", "user_id", "created_at", "updated_at"}
}

func NewTopSong(userID string) TopSong {
	return TopSong{
		ID:     uuid.New(),
		UserID: userID,
	}
}

func (t *TopSong) ToSlice() []interface{} {
	slice := make([]interface{}, 4)
	slice[0] = t.ID
	slice[1] = t.UserID
	slice[2] = utils.Now()
	slice[3] = utils.Now()

	return slice
}
