package subdomain

import (
	"database/sql"
	"fmt"

	"github.com/oligoden/chassis/device/model/data"
)

type Record struct {
	Subdomain string
	URL       string
	data.Default
}

func NewRecord() *Record {
	r := &Record{}
	r.Default = data.Default{}
	r.Perms = "r:::"
	return r
}

func (Record) TableName() string {
	return "subdomains"
}

func (Record) Migrate(db *sql.DB) error {
	q := "CREATE TABLE `subdomains` ( `subdomain` varchar(255) DEFAULT NULL, `url` varchar(255) DEFAULT NULL, `id` int(10) unsigned NOT NULL AUTO_INCREMENT, `uc` varchar(255) DEFAULT NULL, `owner_id` int(10) unsigned DEFAULT NULL, `perms` varchar(255) DEFAULT NULL, `hash` varchar(255) DEFAULT NULL, PRIMARY KEY (`id`), UNIQUE KEY `uc` (`uc`) )"

	_, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("doing migration: %w", err)
	}
	return nil
}
