package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storagePath, migrationsPath, migrationsTableName string
	flag.StringVar(&storagePath, "storage-path", "", "path to storage (for postgres format is: user:password@host:port/dbname?query)")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTableName, "table", "migrations", "name of migrations table")
	flag.Parse()
	if storagePath == "" || migrationsPath == "" {
		panic("storage and migrations paths are required")
	}
	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("postgres://%s?x-migrations-table=%s&sslmode=disable", storagePath, migrationsTableName),
	)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migrations to apply")
			return
		}
		// m.Down()
		panic(err)
	}
	fmt.Println("Migrations applied successfully")
}
