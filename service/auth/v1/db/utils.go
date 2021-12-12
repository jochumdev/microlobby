package db

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"wz2100.net/microlobby/shared/component"
)

type Timestamps struct {
	CreatedAt time.Time    `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at" yaml:"created_at"`
	UpdatedAt bun.NullTime `bun:"updated_at" json:"updated_at" yaml:"updated_at"`
}

type SoftDelete struct {
	DeletedAt bun.NullTime `bun:"deleted_at,soft_delete" json:"deleted_at" yaml:"deleted_at"`
}

func RoleGetId(ctx context.Context, name string) (string, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return "", err
	}

	var result string
	err = bun.NewSelect().Table("roles").Column("id").Limit(1).Where("name = ?", name).Scan(ctx, &result)
	if err != nil || len(result) < 1 {
		return "", fmt.Errorf("role '%s' not found", name)
	}

	return result, nil
}
