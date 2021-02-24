package models

import (
	"time"

	"github.com/google/uuid"
)

type TopArtist struct {
	ID         uuid.UUID `db:"id"`
	ArtistID   string    `db:"artist_id"`
	Order      int       `db:"order"`
	TimePeriod string    `db:"time_period"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func NewTopArtist(name string, artistID string, order int, timePeriod string) TopArtist {
	return TopArtist{
		ID:         uuid.New(),
		ArtistID:   artistID,
		Order:      order,
		TimePeriod: timePeriod,
	}
}

func (a *TopArtist) ToSlice() []interface{} {
	slice := make([]interface{}, 6)
	slice[0] = a.ID
	slice[1] = a.ArtistID
	slice[2] = a.Order
	slice[3] = a.TimePeriod
	slice[4] = time.Now().UTC().Format(time.RFC3339)
	slice[5] = time.Now().UTC().Format(time.RFC3339)

	return slice
}
