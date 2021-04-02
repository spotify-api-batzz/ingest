package models

import (
	"spotify/utils"
	"time"

	"github.com/google/uuid"
)

type TopArtist struct {
	ID     uuid.UUID `db:"id"`
	UserID string    `db:"user_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewTopArtist(userID string) TopArtist {
	return TopArtist{
		ID:     uuid.New(),
		UserID: userID,
	}
}

func (r *TopArtist) TableName() string {
	return "top_artists"
}

func (r *TopArtist) TableColumns() []string {
	return []string{"id", "user_id", "created_at", "updated_at"}
}

func (a *TopArtist) ToSlice() []interface{} {
	slice := make([]interface{}, 4)
	slice[0] = a.ID
	slice[1] = a.UserID
	slice[2] = utils.Now()
	slice[3] = utils.Now()

	return slice
}
