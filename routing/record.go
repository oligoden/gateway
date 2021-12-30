package routing

import (
	"database/sql"
	"fmt"

	"github.com/oligoden/chassis/device/model/data"
)

type Record struct {
	Domain    string
	Path      string
	URL       string
	resetCORS bool
	data.Default
}

func NewRecord() *Record {
	r := &Record{}
	r.Default = data.Default{}
	r.Perms = "r:::"
	return r
}

func (Record) TableName() string {
	return "routings"
}

func (Record) Migrate(db *sql.DB) error {
	q := "CREATE TABLE `routings` ( `domain` varchar(255) DEFAULT NULL, `path` varchar(255) DEFAULT NULL, `url` varchar(255) DEFAULT NULL, `reset_cors` boolean, `id` int(10) unsigned NOT NULL AUTO_INCREMENT, `uc` varchar(255) DEFAULT NULL, `owner_id` int(10) unsigned DEFAULT NULL, `perms` varchar(255) DEFAULT NULL, `hash` varchar(255) DEFAULT NULL, PRIMARY KEY (`id`), UNIQUE KEY `uc` (`uc`) )"

	_, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("doing migration: %w", err)
	}
	return nil
}
