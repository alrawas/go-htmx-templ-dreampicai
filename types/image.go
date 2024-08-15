package types

import (
	"time"

	"github.com/google/uuid"
)

type ImageStatus int

const (
	ImageStatusFailed ImageStatus = iota
	ImageStatusPending
	ImageStatusCompleted
)

type Image struct {
	ID            int `bun:"id,pk,autoincrement"`
	UserID        uuid.UUID
	Status        ImageStatus
	ImageLocation string // d32d32d23d2 to fetch the actual image
	BatchID       uuid.UUID
	Prompt        string
	Deleted       bool      `bun:"default:'false'"`
	CreatedAt     time.Time `bun:"default:'now()'"`
	DeletedAt     time.Time
}
