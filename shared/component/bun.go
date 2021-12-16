package component

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/log/logrusadapter"
	"github.com/jackc/pgx/stdlib"
	"github.com/urfave/cli/v2"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type CBUNKey struct{}

type CBUN struct {
	SQLDB *sql.DB
	Bun   *bun.DB

	initialized bool
}

// NewBUN creates a new component
func NewBUN() *CBUN {
	return &CBUN{initialized: false}
}

func BunFromContext(ctx context.Context) (*bun.DB, error) {
	reg, err := RegistryFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bun, err := Bun(reg)
	if err != nil {
		return nil, err
	}

	return bun, nil
}

func Bun(reg *Registry) (*bun.DB, error) {
	bunc, err := reg.Get(CBUNKey{})
	if err != nil {
		return nil, err
	}

	// bunc.Init(serveRegistry, tt.cmd)
	bun := bunc.(*CBUN).GetBUN()
	if bun == nil {
		return nil, errors.New("bun is nil")
	}

	return bun, nil
}

func (c *CBUN) Priority() int8 {
	return 20
}

func (c *CBUN) Key() interface{} {
	return CBUNKey{}
}

func (c *CBUN) Name() string {
	return "shared.bun"
}

func (c *CBUN) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "database-url",
			Usage:   "bun Database URL",
			EnvVars: []string{"DATABASE_URL"},
		},
		&cli.BoolFlag{
			Name:        "database-debug",
			Usage:       "Set it to the debug the database queries",
			EnvVars:     []string{"DATABASE_DEBUG"},
			DefaultText: "false",
			Value:       false,
		},
		&cli.StringFlag{
			Name:    "migrations-table",
			Value:   "schema_migrations",
			Usage:   "Table to store migrations info",
			EnvVars: []string{"MIGRATIONS_TABLE"},
		},
		&cli.StringFlag{
			Name:    "migrations-dir",
			Value:   "/migrations",
			Usage:   "Folder which contains migrations",
			EnvVars: []string{"MIGRATIONS_DIR"},
		},
	}
}

func (c *CBUN) Initialized() bool {
	return c.initialized
}

func (c *CBUN) Init(registry *Registry, cli *cli.Context) error {
	if c.initialized {
		return nil
	}

	config, err := pgx.ParseURI(cli.String("database-url"))
	if err != nil {
		return err
	}

	config.PreferSimpleProtocol = true

	logrusc, err := Logrus(registry)
	if err == nil {
		config.Logger = logrusadapter.NewLogger(logrusc.Logger())
	}

	c.SQLDB = stdlib.OpenDB(config)
	driver, err := postgres.WithInstance(c.SQLDB, &postgres.Config{MigrationsTable: cli.String("migrations-table")})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", cli.String("migrations-dir")),
		"postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != migrate.ErrNoChange && err != nil {
		return err
	}

	c.Bun = bun.NewDB(c.SQLDB, pgdialect.New())
	if c.Bun == nil {
		return errors.New("failed to create bun")
	}

	if cli.Bool("database-debug") {
		// Print all queries to stdout.
		c.Bun.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	c.initialized = true

	return nil
}

func (c *CBUN) Health(context context.Context) (string, bool) {
	if !c.Initialized() {
		return "Not initialized", true
	}

	if err := c.Bun.Ping(); err != nil {
		return err.Error(), true
	}

	return "Pong received", false
}

func (c *CBUN) Close() error {
	err := c.SQLDB.Close()
	if err != nil {
		return err
	}

	err = c.Bun.Close()
	if err != nil {
		return err
	}

	return nil
}

func (c *CBUN) GetBUN() *bun.DB {
	return c.Bun
}
