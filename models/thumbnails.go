package models

import (
	"fmt"
	"spotify/utils"
)

type Thumbnail struct {
	ID        string     `db:"id"`
	Entity    string     `db:"entity_type"`
	EntityID  string     `db:"entity_id"`
	URL       string     `db:"url"`
	Width     int        `db:"width"`
	Height    int        `db:"height"`
	CreatedAt utils.Time `db:"created_at"`
	UpdatedAt utils.Time `db:"updated_at"`
}

func NewThumbnail(entity string, entityID string, URL string, height float64, width float64) Thumbnail {
	return Thumbnail{

		ID:        utils.GenerateUUID(),
		Entity:    entity,
		EntityID:  entityID,
		Width:     int(width),
		Height:    int(height),
		URL:       URL,
		CreatedAt: utils.NewTime(),
		UpdatedAt: utils.NewTime(),
	}
}

func (r *Thumbnail) TableName() string {
	return "thumbnails"
}

func (r *Thumbnail) UniqueID() string {
	return fmt.Sprintf("%s-%d-%d", r.EntityID, r.Width, r.Height)
}
