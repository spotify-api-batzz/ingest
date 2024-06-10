package models

import (
	"spotify/utils"
)

type Song struct {
	ID        string     `db:"id"`
	SpotifyID string     `db:"spotify_id"`
	AlbumID   string     `db:"album_id"`
	ArtistID  string     `db:"artist_id"`
	Name      string     `db:"name"`
	CreatedAt utils.Time `db:"created_at"`
	UpdatedAt utils.Time `db:"updated_at"`

	NeedsUpdate bool

	Album
}

func (r Song) Identifier() string {
	return r.ID
}

func NewSong(name string, spotifyID string, albumID string, artistID string, needsUpdate bool) Song {
	return Song{
		ID:        utils.GenerateUUID(),
		Name:      name,
		SpotifyID: spotifyID,
		ArtistID:  artistID,
		AlbumID:   albumID,
		CreatedAt: utils.NewTime(),
		UpdatedAt: utils.NewTime(),

		NeedsUpdate: needsUpdate,
	}
}

func (r *Song) TableName() string {
	return "songs"
}
