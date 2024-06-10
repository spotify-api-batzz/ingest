package models

import (
	"spotify/utils"
)

type Artist struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	SpotifyID string     `db:"spotify_id"`
	CreatedAt utils.Time `db:"created_at"`
	UpdatedAt utils.Time `db:"updated_at"`

	NeedsUpdate bool
}

func NewArtist(name string, spotifyID string, needsUpdate bool) Artist {
	return Artist{
		ID:          utils.GenerateUUID(),
		Name:        name,
		SpotifyID:   spotifyID,
		CreatedAt:   utils.NewTime(),
		UpdatedAt:   utils.NewTime(),
		NeedsUpdate: needsUpdate,
	}
}

func (r *Artist) Identifier() string {
	return r.ID
}

func (r *Artist) TableName() string {
	return "artists"
}
