package internal

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2/queue"
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
	_ queue.Storage = (*SqliteStorage)(nil)
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
	statement, err := s.dbh.Prepare("INSERT INTO visited (requestID, visited) VALUES (?, 1)")
	if err != nil {
		return err
	}
	_, err = statement.Exec(int64(requestID))
	if err != nil {
		return err
	}
	return nil
}

func (s *SqliteStorage) IsVisited(requestID uint64) (bool, error) {
	var count int
	statement, err := s.dbh.Prepare("SELECT COUNT(*) FROM visited where requestId = ?")
	if err != nil {
		return false, err
	}
	row := statement.QueryRow(int64(requestID))
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}
	if count >= 1 {
		return true, nil
	}
	return false, nil
}

func (s *SqliteStorage) SetCookies(u *url.URL, cookies string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	statement, err := s.dbh.Prepare("INSERT INTO cookies (host, cookies) VALUES (?, ?)")
	if err != nil {
		log.Printf("SetCookies() .Set error %s", err)
	}
	_, err = statement.Exec(u.Host, cookies)
	if err != nil {
		log.Printf("SetCookies() .Set error %s", err)
	}
}

func (s *SqliteStorage) Cookies(u *url.URL) string {
	var cookies string
	s.mu.RLock()
	defer s.mu.RUnlock()

	statement, err := s.dbh.Prepare("SELECT cookies FROM cookies where host = ?")
	if err != nil {
		log.Printf("Cookies() .Get error %s", err)
		return ""
	}
	row := statement.QueryRow(u.Host)
	err = row.Scan(&cookies)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return ""
		}
		log.Printf("Cookies() .Get error %s", err)
	}
	return cookies
}

func (s *SqliteStorage) AddRequest(b []byte) error {
	statement, err := s.dbh.Prepare("INSERT INTO queue (data) VALUES (?)")
	if err != nil {
		return err
	}
	_, err = statement.Exec(b)
	if err != nil {
		return err
	}
	return nil
}

func (s *SqliteStorage) GetRequest() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var blob []byte
	var id int
	statement, err := s.dbh.Prepare("SELECT min(id), data FROM queue")
	if err != nil {
		return nil, err
	}
	row := statement.QueryRow()
	err = row.Scan(&id, &blob)
	if err != nil {
		return nil, err
	}

	statement, err = s.dbh.Prepare("DELETE FROM queue where id = ?")
	_, err = statement.Exec(id)
	if err != nil {
		return nil, err
	}

	return blob, nil
}

func (s *SqliteStorage) QueueSize() (int, error) {
	var count int
	statement, err := s.dbh.Prepare("SELECT COUNT(*) FROM queue")
	if err != nil {
		return 0, err
	}
	row := statement.QueryRow()
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}