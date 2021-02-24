package models

import (
	"time"

	"github.com/google/uuid"
)

type Album struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	ArtistID  string    `db:"artist_id"`
	SpotifyID string    `db:"spotify_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Artist
}

func NewAlbum(name string, artistID string, spotifyID string) Album {
	return Album{
		ID:        uuid.New(),
		Name:      name,
		ArtistID:  artistID,
		SpotifyID: spotifyID,
	}
}

func (a *Album) ToSlice() []interface{} {
	slice := make([]interface{}, 6)
	slice[0] = a.ID
	slice[1] = a.Name
	slice[2] = a.ArtistID
	slice[3] = a.SpotifyID
	slice[4] = time.Now().UTC().Format(time.RFC3339)
	slice[5] = time.Now().UTC().Format(time.RFC3339)

	return slice
}
