package models

import (
	"spotify/utils"
)

type Album struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	ArtistID  string     `db:"artist_id"`
	SpotifyID string     `db:"spotify_id"`
	CreatedAt utils.Time `db:"created_at"`
	UpdatedAt utils.Time `db:"updated_at"`

	NeedsUpdate bool

	Artist
}

func NewAlbum(name string, artistID string, spotifyID string, needsUpdate bool) Album {
	return Album{
		ID:          utils.GenerateUUID(),
		Name:        name,
		ArtistID:    artistID,
		SpotifyID:   spotifyID,
		CreatedAt:   utils.NewTime(),
		UpdatedAt:   utils.NewTime(),
		NeedsUpdate: needsUpdate,
	}
}

func (r *Album) Identifier() string {
	return r.ID
}

func (r *Album) TableName() string {
	return "albums"
}
