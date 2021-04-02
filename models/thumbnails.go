package models

import (
	"spotify/utils"
	"time"

	"github.com/google/uuid"
)

type Thumbnail struct {
	ID        uuid.UUID `db:"id"`
	Entity    string    `db:"entity"`
	EntityID  string    `db:"entity_id"`
	URL       string    `db:"url"`
	Width     int       `db:"width"`
	Height    int       `db:"height"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewThumbnail(entity string, entityID string, URL string, height int, width int) Thumbnail {
	return Thumbnail{
		ID:       uuid.New(),
		Entity:   entity,
		EntityID: entityID,
		Width:    width,
		Height:   height,
		URL:      URL,
	}
}

func (t *Thumbnail) ToSlice() []interface{} {
	slice := make([]interface{}, 8)
	slice[0] = t.ID
	slice[1] = t.Entity
	slice[2] = t.EntityID
	slice[3] = t.URL
	slice[4] = t.Width
	slice[5] = t.Height
	slice[6] = utils.Now()
	slice[7] = utils.Now()

	return slice
}
