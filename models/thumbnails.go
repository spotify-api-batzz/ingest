package models

import (
	"fmt"
	"spotify/utils"
	"time"

	"github.com/google/uuid"
)

type Thumbnail struct {
	ID        uuid.UUID `db:"id"`
	Entity    string    `db:"entity_type"`
	EntityID  string    `db:"entity_id"`
	URL       string    `db:"url"`
	Width     int       `db:"width"`
	Height    int       `db:"height"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewThumbnail(entity string, entityID string, URL string, height float64, width float64) Thumbnail {
	return Thumbnail{
		ID:       uuid.New(),
		Entity:   entity,
		EntityID: entityID,
		Width:    int(width),
		Height:   int(height),
		URL:      URL,
	}
}

func (r *Thumbnail) TableName() string {
	return "thumbnails"
}

func (r *Thumbnail) TableColumns() []string {
	return []string{"id", "entity_type", "entity_id", "url", "width", "height", "created_at", "updated_at"}
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

func (r *Thumbnail) UniqueID() string {
	return fmt.Sprintf("%s-%d-%d", r.EntityID, r.Width, r.Height)
}
