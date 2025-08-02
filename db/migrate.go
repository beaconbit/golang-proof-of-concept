package db

import (
    "database/sql"
    _ "modernc.org/sqlite"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    "graphite/publisher/config"
)


func RunMigrations() {
    logger := logPrefix("RunMigrations")
    logger.Println("starting migrations")

    conf, err := config.LoadConfigWithDefault()
    if err != nil {
	logger.Fatalf("load config: %v", err)
    }

    db, err := sql.Open(conf.Process.DBType, conf.Process.DBPath)
    if err != nil {
	logger.Fatalf("open db: %v", err)
    }
    defer db.Close()

    driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
    if err != nil {
        logger.Fatalf("create driver: %v", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
	conf.Process.DBMigrationUrl,
        conf.Process.DBDriver, 
        driver,
    )
    if err != nil {
        logger.Fatalf("new migrate instance: %v", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        logger.Fatalf("apply migrations: %v", err)
    }
}

