package session

import (
	"database/sql"
	"fmt"

	"github.com/oligoden/chassis/device/model/data"
)

type Record struct {
	IP        string
	UserAgent string
	data.Default
}

func NewRecord() *Record {
	r := &Record{}
	r.Default = data.Default{}
	r.Perms = ":::cr"
	return r
}

func (Record) TableName() string {
	return "sessions"
}

func (e *Record) IDValue(id ...uint) uint {
	if len(id) > 0 {
		e.ID = id[0]
	}
	return e.ID
}

func (Record) Migrate(db *sql.DB) error {
	q := "CREATE TABLE `sessions` (`ip` varchar(255), `user_agent` varchar(511), `id` int unsigned AUTO_INCREMENT, `uc` varchar(255) UNIQUE, `owner_id` int unsigned, `perms` varchar(255), `hash` varchar(255), PRIMARY KEY (`id`))"

	_, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("doing migration: %w", err)
	}
	return nil
}

type SessionUsersRecord struct {
	SessionID uint `json:"-"`
	UserID    uint `json:"-"`
	data.Default
}

func NewSessionUsersRecord() *SessionUsersRecord {
	r := &SessionUsersRecord{}
	r.Default = data.Default{}
	r.Perms = ":::crd"
	return r
}

func (SessionUsersRecord) TableName() string {
	return "session_users"
}

func (SessionUsersRecord) Migrate(db *sql.DB) error {
	q := "CREATE TABLE `session_users` (`session_id` int unsigned, `user_id` int unsigned, `id` int unsigned AUTO_INCREMENT, `uc` varchar(255) UNIQUE, `owner_id` int unsigned, `perms` varchar(255), `hash` varchar(255), PRIMARY KEY (`id`))"
	_, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("doing migration: %w", err)
	}

	return nil
}
