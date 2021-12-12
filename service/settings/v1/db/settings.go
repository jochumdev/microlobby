package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"wz2100.net/microlobby/shared/component"
)

type Setting struct {
	bun.BaseModel `bun:"settings,alias:s"`
	ID            uuid.UUID `bun:"id,type:uuid" json:"id" yaml:"id"`
	OwnerID       uuid.UUID `bun:"owner_id,type:uuid" json:"owner_id" yaml:"owner_id"`
	Service       string    `json:"service" yaml:"service"`
	Name          string    `json:"name" yaml:"name"`
	Content       string    `json:"content" yaml:"content"`
	RolesRead     []string  `bun:"roles_read,array" json:"roles_read" yaml:"roles_read"`
	RolesUpdate   []string  `bun:"roles_update,array" json:"roles_update" yaml:"roles_update"`

	Timestamps
	SoftDelete
}

func (s *Setting) UserHasReadPermission(ctx context.Context) {

}

func (s *Setting) UserHasWritePermission(ctx context.Context) {
}

func SettingsGet(ctx context.Context, id, ownerID, service, name string) (*Setting, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var result Setting
	sql := bun.NewSelect().
		Model(&result).
		ColumnExpr("s.*").
		Limit(1)

	if len(id) > 1 {
		sql.Where("id = ?", id)
	} else if len(ownerID) > 1 {
		if len(name) > 1 {
			sql.Where("owner_id = ? AND name = ?", ownerID, name)
		} else {
			sql.Where("owner_id = ?", ownerID)
		}
	} else if len(service) > 1 {
		if len(name) > 1 {
			sql.Where("service = ? AND name = ?", service, name)
		} else {
			sql.Where("service = ?", service)
		}
	} else {
		return nil, errors.New("not enough parameters")
	}

	err = sql.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func SettingsList(ctx context.Context, id, ownerID, service, name string, limit, offset uint64) ([]Setting, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the data from the db.
	var settings []Setting
	sql := bun.NewSelect().
		Model(&settings).
		ColumnExpr("s.*").
		Limit(int(limit)).
		Offset(int(offset))

	if len(id) > 1 {
		sql.Where("id = ?", id)
	} else if len(ownerID) > 1 {
		if len(name) > 1 {
			sql.Where("owner_id = ? AND name = ?", ownerID, name)
		} else {
			sql.Where("owner_id = ?", ownerID)
		}
	} else if len(service) > 1 {
		if len(name) > 1 {
			sql.Where("service = ? AND name = ?", service, name)
		} else {
			sql.Where("service = ?", service)
		}
	}

	err = sql.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return settings, nil
}
