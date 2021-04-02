package models

import (
	"spotify/utils"
	"time"

	"github.com/google/uuid"
)

type TopSongData struct {
	ID         uuid.UUID `db:"id"`
	TopSongID  string    `db:"top_song_id"`
	SongID     string    `db:"song_id"`
	Order      int       `db:"order"`
	TimePeriod string    `db:"time_period"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`

	Song
	User
}

func (r *TopSongData) TableName() string {
	return "top_song_data"
}

func (r *TopSongData) TableColumns() []string {
	return []string{"id", "top_song_id", "song_id", `"order"`, "time_period", "created_at", "updated_at"}
}

func NewTopSongData(topSongID string, songID string, order int, timePeriod string) TopSongData {
	return TopSongData{
		ID:         uuid.New(),
		SongID:     songID,
		TopSongID:  topSongID,
		Order:      order,
		TimePeriod: timePeriod,
	}
}

func (t *TopSongData) ToSlice() []interface{} {
	slice := make([]interface{}, 7)
	slice[0] = t.ID
	slice[1] = t.TopSongID
	slice[2] = t.SongID
	slice[3] = t.Order
	slice[4] = t.TimePeriod
	slice[5] = utils.Now()
	slice[6] = utils.Now()

	return slice
}
