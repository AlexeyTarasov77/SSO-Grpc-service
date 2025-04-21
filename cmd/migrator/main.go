package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"sso.service/internal/config"
)

func main() {
	var storagePath, migrationsPath string
	flag.StringVar(&storagePath, "storage-path", "", "path to storage (e.g: schema://user:password@host:port/dbname)")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.Parse()
	if migrationsPath == "" {
		currDir, err := os.Getwd()
		if err != nil {
			panic("Unable to resolve migrations location")
		}
		migrationsPath = path.Join(currDir, "migrations")
		if _, err := os.Stat(migrationsPath); err != nil {
			panic(err)
		}
	}
	if storagePath == "" {
		cfgPath := config.ResolveConfigPath()
		cfg := config.MustLoad(cfgPath)
		storagePath = cfg.DB.Dsn
	}
	m, err := migrate.New("file://"+migrationsPath, storagePath)
	if err != nil {
		panic(err)
	}
	var direction string
	if len(os.Args) > 1 {
		direction = os.Args[1]
	} else {
		direction = "up"
	}
	switch strings.ToLower(direction) {
	case "up":
		if migrated := migrateUp(m); !migrated {
			return
		}
	case "down":
		var steps string
		if len(os.Args) > 2 {
			steps = os.Args[2]
		}
		if migrated := migrateDown(steps, m); !migrated {
			return
		}
	default:
		panic(fmt.Sprintf("Unknown direction: %s", direction))
	}
	fmt.Println("Migrations applied successfully")
}

func migrateUp(m *migrate.Migrate) (migrated bool) {
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migrations to apply")
			return
		}
		if fixed := checkAndFixDirty(err, m); !fixed {
			panic(err)
		}
		return
	}
	migrated = true
	return
}

func migrateDown(steps string, m *migrate.Migrate) (migrated bool) {
	if steps != "" {
		stepsInt, err := strconv.Atoi(steps)
		if err != nil {
			panic(err)
		}
		if err := m.Steps(stepsInt); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No migrations to apply")
				return
			}
			if fixed := checkAndFixDirty(err, m); !fixed {
				panic(err)
			}
			return
		}
		migrated = true
		return
	}
	fmt.Print("Are you sure that you want to downgrade all migrations? (yes/no) [no] ")
	var confirmed string
	if _, err := fmt.Scanln(&confirmed); err != nil {
		panic(err)
	}
	if strings.ToLower(confirmed) == "no" {
		return
	}
	if err := m.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migrations to apply")
			return
		}
		if fixed := checkAndFixDirty(err, m); !fixed {
			panic(err)
		}
		return
	}
	migrated = true
	return
}

func checkAndFixDirty(err error, m *migrate.Migrate) (isDirty bool) {
	var dirtyErr migrate.ErrDirty
	if !errors.As(err, &dirtyErr) {
		return
	}
	if err := m.Force(dirtyErr.Version - 1); err != nil {
		panic(fmt.Sprintf("Error during recovering dirty error: %s", err))
	}
	isDirty = true
	fmt.Printf("Fixed dirty migrations error. Try again please")
	return
}
