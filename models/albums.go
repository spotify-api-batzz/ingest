package models

import (
	"spotify/utils"
	"time"

	"github.com/google/uuid"
)

type Album struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	ArtistID  string    `db:"artist_id"`
	SpotifyID string    `db:"spotify_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	NeedsUpdate bool

	Artist
}

func NewAlbum(name string, artistID string, spotifyID string) Album {
	return Album{
		ID:        uuid.New().String(),
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
	slice[4] = utils.Now()
	slice[5] = utils.Now()

	return slice
}
