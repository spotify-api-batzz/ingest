package models

import (
	"time"

	"github.com/google/uuid"
)

type Song struct {
	ID        uuid.UUID `db:"id"`
	SpotifyID string    `db:"spotify_id"`
	AlbumID   string    `db:"album_id"`
	ArtistID  string    `db:"artist_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Album
}

func NewSong(name string, spotifyID string, albumID string, artistID string) Song {
	return Song{
		ID:        uuid.New(),
		Name:      name,
		SpotifyID: spotifyID,
		ArtistID:  artistID,
		AlbumID:   albumID,
	}
}

func (s *Song) ToSlice() []interface{} {
	slice := make([]interface{}, 7)
	slice[0] = s.ID
	slice[1] = s.SpotifyID
	slice[2] = s.AlbumID
	slice[3] = s.ArtistID
	slice[4] = s.Name
	slice[5] = time.Now().UTC().Format(time.RFC3339)
	slice[6] = time.Now().UTC().Format(time.RFC3339)

	return slice
}
