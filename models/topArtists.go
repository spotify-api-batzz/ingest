package models

import (
	"time"

	"github.com/google/uuid"
)

type TopArtist struct {
	ID       uuid.UUID `db:"id"`
	UserID   string    `db:"user_id"`
	ArtistID string    `db:"artist_id"`
	Order    int       `db:"order"`

	TimePeriod string    `db:"time_period"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func NewTopArtist(name string, artistID string, order int, timePeriod string, userID string) TopArtist {
	return TopArtist{
		ID:         uuid.New(),
		ArtistID:   artistID,
		UserID:     userID,
		Order:      order,
		TimePeriod: timePeriod,
	}
}

func (a *TopArtist) ToSlice() []interface{} {
	slice := make([]interface{}, 7)
	slice[0] = a.ID
	slice[1] = a.ArtistID
	slice[2] = a.Order
	slice[3] = a.UserID
	slice[4] = a.TimePeriod
	slice[5] = time.Now().UTC().Format(time.RFC3339)
	slice[6] = time.Now().UTC().Format(time.RFC3339)

	return slice
}
