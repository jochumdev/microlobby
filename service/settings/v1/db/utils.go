package db

import (
	"time"

	"github.com/uptrace/bun"
)

type Timestamps struct {
	CreatedAt time.Time    `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at" yaml:"created_at"`
	UpdatedAt bun.NullTime `bun:"updated_at" json:"updated_at" yaml:"updated_at"`
}

type SoftDelete struct {
	DeletedAt bun.NullTime `bun:"deleted_at,soft_delete" json:"deleted_at" yaml:"deleted_at"`
}
