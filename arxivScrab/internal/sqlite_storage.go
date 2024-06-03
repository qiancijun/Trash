package internal

import (
	"database/sql"
	"fmt"
	"net/url"
	"sync"

	"github.com/gocolly/colly/v2/storage"
	"github.com/qiancijun/Trash/arxivScrab/util"
)

// https://gitcode.com/velebak/colly-sqlite3-storage/blob/master/colly/sqlite3/sqlite3.go

type SqliteStorage struct {
	dbh *sql.DB
	mu sync.RWMutex
}

var (
	_ storage.Storage = (*SqliteStorage)(nil)
)

func (s *SqliteStorage) Init() error {
	if s.dbh == nil {
		db, err := sql.Open("sqlite3", util.RootPath + "data/scrab.sqlite")
		if err != nil {
			return fmt.Errorf("unable to open db file: %s", err)
		}
		err = db.Ping()
		if err != nil {
			return fmt.Errorf("db init failure: %s", err)
		}
		s.dbh = db
	}

	statement, _ := s.dbh.Prepare("CREATE TABLE IF NOT EXISTS visited (id INTEGER PRIMARY KEY, requestID INTEGER, visited INT)")
	_, err := statement.Exec()
	if err != nil {
		return err
	}
	statement, _ = s.dbh.Prepare("CREATE INDEX IF NOT EXISTS idx_visited ON visited (requestID)")
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	statement, _ = s.dbh.Prepare("CREATE TABLE IF NOT EXISTS cookies (id INTEGER PRIMARY KEY, host TEXT, cookies TEXT)")
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	statement, err = s.dbh.Prepare("CREATE INDEX IF NOT EXISTS idx_cookies ON cookies (host)")
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	statement, err = s.dbh.Prepare("CREATE TABLE IF NOT EXISTS queue (id INTEGER PRIMARY KEY, data BLOB)")
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	return nil
}

func (s *SqliteStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	statement, err := s.dbh.Prepare("DROP TABLE visited")
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	statement, err = s.dbh.Prepare("DROP TABLE cookies")
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	statement, err = s.dbh.Prepare("DROP TABLE queue")
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	return nil
}

func (s *SqliteStorage) Visited(requestID uint64) error {
	return nil
}

func (s *SqliteStorage) IsVisited(requestID uint64) (bool, error) {
	return false, nil
}

func (s *SqliteStorage) SetCookies(u *url.URL, cookies string) {

}

func (s *SqliteStorage) Cookies(u *url.URL) string {
	return ""
}