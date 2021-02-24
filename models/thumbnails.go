package models

import "time"

type Thumbnail struct {
	ID        string    `db:"id"`
	Entity    string    `db:"entity"`
	EntityID  string    `db:"entity_id"`
	URL       string    `db:"url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
