package models

import (
	"spotify/utils"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        string    `db:"id"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	SpotifyID string    `db:"spotify_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewUser(username string, password string, spotifyID string) User {
	return User{
		ID:        uuid.New().String(),
		SpotifyID: spotifyID,
		Username:  username,
		Password:  password,
	}
}

func (u *User) ToSlice() []interface{} {
	slice := make([]interface{}, 6)
	slice[0] = u.ID
	slice[1] = u.Username
	slice[2] = u.Password
	slice[3] = u.SpotifyID
	slice[4] = utils.Now()
	slice[5] = utils.Now()

	return slice
}
