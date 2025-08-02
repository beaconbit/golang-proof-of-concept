#cd db/
#./create_migration create -ext sql -dir migrations create_users_table
#
#go mod tidy
#
#go run main


mkdir graphite-publisher && cd graphite-publisher
go mod init graphite/publisher

go get gorm.io/gorm
go get gorm.io/driver/sqlite
go get modernc.org/sqlite
go get github.com/nats-io/nats.go
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/sqlite3
go get github.com/golang-migrate/migrate/v4/source/file

# If you are really creating this project from scrathc you'll need this to run migrate commands from cli
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate.linux-amd64 /usr/local/bin/migrate

mkdir db
touch db/models.go
touch db/migrate.go

echo "
package db

type User struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string
    Email string
}
" > db/models.go

echo "
package db

import (
    "database/sql"
    "log"

    _ "modernc.org/sqlite"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations() {
    db, _ := sql.Open("sqlite", "./data/app.db")

    driver, _ := sqlite3.WithInstance(db, &sqlite3.Config{})
    m, _ := migrate.NewWithDatabaseInstance("file://db/migrations", "sqlite3", driver)
    _ = m.Up()
}" > db/migrate.go

mkdir -p db/migrations

# Note that "-ext sql" sets the file extension of the migration files to .sql
migrate create -ext sql -dir db/migrations init_schema

echo "
CREATE TABLE event_sources (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT,
  ip TEXT,
  port INTEGER
);

CREATE TABLE tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_source_id INTEGER,
  key TEXT,
  value TEXT
); " > db/migrations/000001_init_schema.up.sql

echo "
INSERT INTO event_sources (name, ip, port) VALUES
  ('sensor1', '192.168.1.10', 502),
  ('sensor2', '192.168.1.11', 502);

INSERT INTO tags (event_source_id, key, value) VALUES
  (1, 'location', 'factory-floor'),
  (2, 'location', 'warehouse');

" > db/migrations/000002_seed_data.up.sql
echo " 
DELETE FROM tags;
DELETE FROM event_sources;
" > db/migrations/000002_seed_data.down.sql

go mod tidy
# golang-proof-of-concept
