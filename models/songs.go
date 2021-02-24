package models

import (
	"time"

	"github.com/google/uuid"
)

type Song struct {
	ID        uuid.UUID `db:"id"`
	SpotifyID string    `db:"spotify_id"`
	AlbumID   string    `db:"album_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Album
}

func NewSong(name string, spotifyID string, albumID string) Song {
	return Song{
		ID:        uuid.New(),
		Name:      name,
		SpotifyID: spotifyID,
		AlbumID:   albumID,
	}
}

func (s *Song) ToSlice() []interface{} {
	slice := make([]interface{}, 6)
	slice[0] = s.ID
	slice[1] = s.SpotifyID
	slice[2] = s.AlbumID
	slice[3] = s.Name
	slice[4] = time.Now().UTC().Format(time.RFC3339)
	slice[5] = time.Now().UTC().Format(time.RFC3339)

	return slice
}
