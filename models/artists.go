package models

import (
	"time"

	"github.com/google/uuid"
)

type Artist struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	SpotifyID string    `db:"spotify_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	NeedsUpdate bool
}

func NewArtist(name string, spotifyID string) Artist {
	return Artist{
		ID:        uuid.New().String(),
		Name:      name,
		SpotifyID: spotifyID,
	}
}

func (a *Artist) ToSlice() []interface{} {
	slice := make([]interface{}, 5)
	slice[0] = a.ID
	slice[1] = a.Name
	slice[2] = a.SpotifyID
	slice[3] = time.Now().In(time.UTC).Format(time.RFC3339)
	slice[4] = time.Now().In(time.UTC).Format(time.RFC3339)

	return slice
}
