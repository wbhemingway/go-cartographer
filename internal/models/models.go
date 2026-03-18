package models

import (
	"errors"
	"time"
)

type MapStatus string

const (
	StatusPending   MapStatus = "pending"
	StatusCompleted MapStatus = "completed"
)

type UserRoles string

const (
	RoleAdmin UserRoles = "admin"
	RoleUser  UserRoles = "user"
)

var (
	ErrMapNotFound   = errors.New("map not found")
	ErrUnauthorized  = errors.New("user does not own this map")
	ErrInvalidConfig = errors.New("map configuration is invalid")
	ErrOutOfBounds = errors.New("coordinate falls outside map dimensions")
)

type Tile struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Terrain   string `json:"terrain"`
	Creature  string `json:"creature"`
	Structure string `json:"structure"`
}

type World struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Tiles  []Tile `json:"tiles"`
}

type MapResponse struct {
	ID     string    `json:"id"`
	Status MapStatus `json:"status"`
	URL    string    `json:"url,omitempty"`
}

type MapMetadata struct {
	ID               string    `firestore:"id" json:"id"`
	CreatorID        string    `firestore:"creator_id" json:"creator_id"`
	ConfigObjectName string    `firestore:"config_object_name" json:"config_object_name"`
	CreatedAt        time.Time `firestore:"created_at" json:"created_at"`
	Status           MapStatus `firestore:"status" json:"status"`
}

type APIKey struct {
	ID        string    `firestore:"id"`
	UserID    string    `firestore:"user_id"`
	UserRole  UserRoles `firestore:"user_role"`
	IsActive  bool      `firestore:"is_active"`
	CreatedAt time.Time `firestore:"created_at"`
}

type User struct {
	ID        string    `firestore:"id"`
	Email     string    `firestore:"email"`
	Role      UserRoles `firestore:"role"`
	CreatedAt time.Time `firestore:"created_at"`
}
