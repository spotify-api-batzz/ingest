package models

import "time"

type Album struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	ArtistID  string    `db:"artist_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Artist
}
