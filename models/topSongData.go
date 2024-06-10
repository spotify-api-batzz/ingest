package models

import (
	"spotify/utils"
)

type TopSongData struct {
	ID         string     `db:"id"`
	TopSongID  string     `db:"top_song_id"`
	SongID     string     `db:"song_id"`
	Order      int        `db:"order"`
	TimePeriod string     `db:"time_period"`
	CreatedAt  utils.Time `db:"created_at"`
	UpdatedAt  utils.Time `db:"updated_at"`

	Song
	User
}

func (r *TopSongData) TableName() string {
	return "top_song_data"
}

func NewTopSongData(topSongID string, songID string, order int, timePeriod string) TopSongData {
	return TopSongData{
		ID:         utils.GenerateUUID(),
		SongID:     songID,
		TopSongID:  topSongID,
		Order:      order,
		TimePeriod: timePeriod,
		CreatedAt:  utils.NewTime(),
		UpdatedAt:  utils.NewTime(),
	}
}
