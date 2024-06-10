package models

import (
	"spotify/utils"
)

type TopArtistData struct {
	ID          string `db:"id"`
	ArtistID    string `db:"artist_id"`
	TopArtistID string `db:"top_artist_id"`
	Order       int    `db:"order"`

	TimePeriod string     `db:"time_period"`
	CreatedAt  utils.Time `db:"created_at"`
	UpdatedAt  utils.Time `db:"updated_at"`
}

func NewTopArtistData(name string, artistID string, order int, timePeriod string, topArtistID string) TopArtistData {
	return TopArtistData{
		ID:          utils.GenerateUUID(),
		ArtistID:    artistID,
		TopArtistID: topArtistID,
		Order:       order,
		TimePeriod:  timePeriod,
		CreatedAt:   utils.NewTime(),
		UpdatedAt:   utils.NewTime(),
	}
}

func (r *TopArtistData) TableName() string {
	return "top_artist_data"
}
