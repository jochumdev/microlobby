package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"wz2100.net/microlobby/shared/component"
)

type User struct {
	bun.BaseModel `bun:"users,alias:u"`
	ID            uuid.UUID `bun:"id,type:uuid" json:"id" yaml:"id"`
	Username      string    `json:"username" yaml:"username"`
	Password      string    `json:"-" yaml:"-"`
	Email         string    `json:"email" yaml:"email"`
	Roles         []string  `bun:",array,scanonly" json:"roles" yaml:"roles"`

	Timestamps
	SoftDelete
}

func UserList(ctx context.Context, limit, offset uint64) ([]User, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get the data from the db.
	var users []User
	err = bun.NewSelect().
		Model(&users).
		ColumnExpr("u.*").
		ColumnExpr("array(SELECT r.name FROM users_roles AS ur LEFT JOIN roles AS r ON ur.role_id = r.id WHERE ur.user_id = u.id) AS roles").
		Limit(int(limit)).
		Offset(int(offset)).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func UserDetail(ctx context.Context, id string) (*User, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user := User{}
	err = bun.NewSelect().
		Model(&user).
		ColumnExpr("u.*").
		ColumnExpr("array(SELECT r.name FROM users_roles AS ur LEFT JOIN roles AS r ON ur.role_id = r.id WHERE ur.user_id = u.id) AS roles").
		Limit(1).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UserDelete(ctx context.Context, id string) error {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return err
	}

	user := User{}
	_, err = bun.NewDelete().Model(&user).Where("id = ?", id).Exec(ctx)
	return err
}

func UserUpdateRoles(ctx context.Context, id string, roles []string) (*User, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if all new roles exists
	rolesIds := make([]string, len(roles))
	for idx, role := range roles {
		id, err := RoleGetId(ctx, role)
		if err != nil {
			return nil, err
		}
		rolesIds[idx] = id
	}

	// Delete all current roles
	_, err = bun.NewDelete().Table("users_roles").Where("user_id = ?", id).Exec(ctx)
	if err != nil {
		return nil, err
	}

	// Exit out if user wants to delete all roles
	if len(roles) < 1 {
		return UserDetail(ctx, id)
	}

	// Reassign roles
	for _, roleId := range rolesIds {
		values := map[string]interface{}{
			"user_id": id,
			"role_id": roleId,
		}
		_, err = bun.NewInsert().Model(&values).TableExpr("users_roles").Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	return UserDetail(ctx, id)
}

func UserFindByUsername(ctx context.Context, username string) (*User, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user := User{}
	err = bun.NewSelect().
		Model(&user).
		ColumnExpr("u.*").
		ColumnExpr("array(SELECT r.name FROM users_roles AS ur LEFT JOIN roles AS r ON ur.role_id = r.id WHERE ur.user_id = u.id) AS roles").
		Limit(1).
		Where("username = ?", username).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UserFindById(ctx context.Context, id string) (*User, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user := User{}
	err = bun.NewSelect().
		Model(&user).
		ColumnExpr("u.*").
		ColumnExpr("array(SELECT r.name FROM users_roles AS ur LEFT JOIN roles AS r ON ur.role_id = r.id WHERE ur.user_id = u.id) AS roles").
		Limit(1).
		Where("u.id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UserCreate(ctx context.Context, username, password string, roles []string) (*User, error) {
	// Get the database engine
	bun, err := component.BunFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Create the user
	user := User{}
	user.Username = username
	user.Password = password
	_, err = bun.NewInsert().Model(&user).Exec(ctx, &user)
	if err != nil {
		return nil, err
	}

	// Create roles
	_, err = UserUpdateRoles(ctx, user.ID.String(), roles)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
