package data

import "time"

type Product struct {
	Name         string    `bson:"name"`
	Price        float64   `bson:"price"`
	UpdatedAt    time.Time `bson:"updated_at"`
	UpdatesCount int       `bson:"updates_count"`
}
