package models

import "github.com/google/uuid"

type RatingChange struct {
	UserID    uuid.UUID `json:"user_id"`
	OldRating int       `json:"old_rating"`
	NewRating int       `json:"new_rating"`
	Delta     int       `json:"delta"`
}
