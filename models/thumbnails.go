package models

import (
	"time"

	"github.com/google/uuid"
)

type Thumbnail struct {
	ID        uuid.UUID `db:"id"`
	Entity    string    `db:"entity"`
	EntityID  string    `db:"entity_id"`
	URL       string    `db:"url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewThumbnail(entity string, entityID string, URL string) Thumbnail {
	return Thumbnail{
		ID:       uuid.New(),
		Entity:   entity,
		EntityID: entityID,
		URL:      URL,
	}
}

func (t *Thumbnail) ToSlice() []interface{} {
	slice := make([]interface{}, 6)
	slice[0] = t.ID
	slice[1] = t.Entity
	slice[2] = t.EntityID
	slice[3] = t.URL
	slice[4] = time.Now().UTC().Format(time.RFC3339)
	slice[5] = time.Now().UTC().Format(time.RFC3339)

	return slice
}
