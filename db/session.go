package db

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "graphite/publisher/config"
)

func NewSession() *gorm.DB {
    logger := logPrefix("NewSession")
    logger.Println("Creating new sqlite session")

    conf, err := config.LoadConfigWithDefault()
    if err != nil {
        logger.Fatalf("load config: %v", err)
    }

    db, err := gorm.Open(sqlite.Open(conf.Process.DBPath), &gorm.Config{})
    if err != nil {
        logger.Fatalf("failed to connect database: %v", err)
    }
    return db
}
