package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go-micro.dev/v4/util/log"
	"jochum.dev/jo-micro/auth2"
	"wz2100.net/microlobby/shared/component"
	sdb "wz2100.net/microlobby/shared/db"
	"wz2100.net/microlobby/shared/proto/settingsservicepb/v1"
)

type Setting struct {
	bun.BaseModel `bun:"settings,alias:s"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" yaml:"id"`
	OwnerID       uuid.UUID `bun:"owner_id,type:uuid" json:"owner_id" yaml:"owner_id"`
	Service       string    `json:"service" yaml:"service"`
	Name          string    `json:"name" yaml:"name"`
	Content       []byte    `bun:"content,type:bytea" json:"content" yaml:"content"`
	RolesRead     []string  `bun:"roles_read,array" json:"roles_read" yaml:"roles_read"`
	RolesUpdate   []string  `bun:"roles_update,array" json:"roles_update" yaml:"roles_update"`

	sdb.Timestamps
	sdb.SoftDelete
}

func (s *Setting) UserHasReadPermission(ctx context.Context) bool {
	user, err := auth2.ClientAuthRegistry().Plugin().Inspect(ctx)
	if err != nil {
		log.Error(err)
		return false
	}

	if auth2.HasRole(user, auth2.ROLE_SUPERADMIN) {
		return true
	}

	if user.Id == s.OwnerID.String() {
		return true
	}

	if auth2.IntersectsRoles(user, s.RolesRead...) {
		return true
	}

	return false
}

func (s *Setting) UserHasUpdatePermission(ctx context.Context) bool {
	user, err := auth2.ClientAuthRegistry().Plugin().Inspect(ctx)
	if err != nil {
		return false
	}

	if auth2.HasRole(user, auth2.ROLE_SUPERADMIN) {
		return true
	}

	if user.Id == s.OwnerID.String() {
		return true
	}

	if auth2.IntersectsRoles(user, s.RolesUpdate...) {
		return true
	}

	return false
}

func SettingsCreate(ctx context.Context, in *settingsservicepb.CreateRequest) (*Setting, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var result Setting
	if len(in.OwnerId) > 0 {
		result.OwnerID = uuid.MustParse(in.OwnerId)
	}
	result.Service = in.Service
	result.Name = in.Name
	result.Content = in.Content
	result.RolesRead = in.RolesRead
	result.RolesUpdate = in.RolesUpdate

	_, err = bun.NewInsert().Model(&result).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func SettingsUpdate(ctx context.Context, id, ownerID, service, name string, content []byte) (*Setting, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch current setting
	s, err := SettingsGet(ctx, id, ownerID, service, name)
	if err != nil {
		return nil, err
	}

	if !s.UserHasUpdatePermission(ctx) {
		return nil, errors.New("unauthorized")
	}

	s.Content = content
	s.UpdatedAt.Time = time.Now()

	// Update
	_, err = bun.NewUpdate().Model(s).Where("id = ?", s.ID).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func SettingsUpsert(ctx context.Context, in *settingsservicepb.UpsertRequest) (*Setting, error) {
	s, err := SettingsUpdate(ctx, "", in.OwnerId, in.Service, in.Name, in.Content)
	if err == nil {
		return s, nil
	}

	req := settingsservicepb.CreateRequest{}
	req.Service = in.Service
	req.Name = in.Name
	req.Content = in.Content
	req.RolesRead = in.RolesRead
	req.RolesUpdate = in.RolesUpdate

	return SettingsCreate(ctx, &req)
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
	} else if len(service) > 1 {
		if len(name) > 1 {
			sql.Where("service = ? AND name = ?", service, name)
		} else {
			sql.Where("service = ?", service)
		}
	} else if len(ownerID) > 1 {
		if len(name) > 1 {
			sql.Where("owner_id = ? AND name = ?", ownerID, name)
		} else {
			sql.Where("owner_id = ?", ownerID)
		}
	} else {
		return nil, errors.New("not enough parameters")
	}

	err = sql.Scan(ctx)
	if err != nil {
		return nil, err
	}

	if !result.UserHasReadPermission(ctx) {
		return nil, errors.New("unauthorized")
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
	} else if len(service) > 1 {
		if len(name) > 1 {
			sql.Where("service = ? AND name = ?", service, name)
		} else {
			sql.Where("service = ?", service)
		}
	} else if len(ownerID) > 1 {
		if len(name) > 1 {
			sql.Where("owner_id = ? AND name = ?", ownerID, name)
		} else {
			sql.Where("owner_id = ?", ownerID)
		}
	}

	err = sql.Scan(ctx)
	if err != nil {
		return nil, err
	}

	var result []Setting
	for _, s := range settings {
		if s.UserHasReadPermission(ctx) {
			result = append(result, s)
		}
	}

	return result, nil
}
