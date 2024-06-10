package models

import (
	"spotify/utils"
)

type User struct {
	ID        string     `db:"id"`
	Username  string     `db:"username"`
	Password  string     `db:"password"`
	SpotifyID string     `db:"spotify_id"`
	CreatedAt utils.Time `db:"created_at"`
	UpdatedAt utils.Time `db:"updated_at"`
}

func (r *User) TableName() string {
	return "users"
}

func NewUser(username string, password string, spotifyID string) User {
	return User{
		ID:        utils.GenerateUUID(),
		SpotifyID: spotifyID,
		Username:  username,
		Password:  password,
		CreatedAt: utils.NewTime(),
		UpdatedAt: utils.NewTime(),
	}
}
