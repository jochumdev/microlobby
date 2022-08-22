package db

import (
	"context"
	"fmt"

	"wz2100.net/microlobby/shared/component"
)

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
