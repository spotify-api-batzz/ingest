package models

import (
	"spotify/utils"
	"time"

	"github.com/google/uuid"
)

type TopSong struct {
	ID         uuid.UUID `db:"id"`
	UserID     string    `db:"user_id"`
	SongID     string    `db:"song_id"`
	Order      int       `db:"order"`
	TimePeriod string    `db:"time_period"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`

	Song
	User
}

func NewTopSong(userID string, songID string, order int, timePeriod string) TopSong {
	return TopSong{
		ID:         uuid.New(),
		SongID:     songID,
		UserID:     userID,
		Order:      order,
		TimePeriod: timePeriod,
	}
}

func (t *TopSong) ToSlice() []interface{} {
	slice := make([]interface{}, 7)
	slice[0] = t.ID
	slice[1] = t.UserID
	slice[2] = t.SongID
	slice[3] = t.Order
	slice[4] = t.TimePeriod
	slice[5] = utils.Now()
	slice[6] = utils.Now()

	return slice
}
