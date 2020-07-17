package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/visiperf/visisql/v2"

	_ "github.com/lib/pq"
)

const PG_USER = "PG_USER"
const PG_PWD = "PG_PASSWORD"
const PG_HOST = "localhost"
const PG_PORT = 5432
const PG_DB_NAME = "PG_DB_NAME"
const PG_OPTIONS = "sslmode=disable"

const TABLE_NAME = "YOUR_TABLE_NAME"

type Site struct {
	ID    int64  `sql:"id"`
	URL   string `sql:"url"`
	Image string `sql:"image"`
}

func (s *Site) String() string {
	return fmt.Sprintf("id: %d - url: %s - image: %s", s.ID, s.URL, s.Image)
}

func main() {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s", PG_USER, PG_PWD, PG_HOST, PG_PORT, PG_DB_NAME, PG_OPTIONS))
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	id, err := insert(db)
	if err != nil {
		log.Fatalf("failed to insert: %v", err)
	}

	s, err := get(db, id)
	if err != nil {
		log.Fatalf("failed to get: %v", err)
	}
	fmt.Println(s)

	if err := update(db, id); err != nil {
		log.Fatalf("failed to update: %v", err)
	}

	s, err = get(db, id)
	if err != nil {
		log.Fatalf("failed to get: %v", err)
	}
	fmt.Println(s)

	if err := remove(db, id); err != nil {
		log.Fatalf("failed to delete: %v", err)
	}

	if err := insertMultiple(db); err != nil {
		log.Fatalf("failed to insert multiple: %v", err)
	}

	ss, err := list(db)
	if err != nil {
		log.Fatalf("failed to list: %v", err)
	}
	fmt.Println(ss)
}

func insert(db *sql.DB) (interface{}, error) {
	ts, err := visisql.NewTransactionService(db)
	if err != nil {
		return nil, fmt.Errorf("insert failed to init transaction: %w", err)
	}

	data := map[string]interface{}{
		"url":   "mUrl",
		"image": "mImage",
	}

	id, err := ts.Insert(TABLE_NAME, data, "id")
	if err != nil {
		return nil, fmt.Errorf("insert failed to insert data: %w", err)
	}

	if err := ts.Commit(); err != nil {
		return nil, fmt.Errorf("insert failed to commit transaction: %w", err)
	}

	return id, nil
}

func get(db *sql.DB, id interface{}) (*Site, error) {
	predicates := [][]*visisql.Predicate{{
		visisql.NewPredicate("id", visisql.OperatorEqual, []interface{}{id}),
	}}

	var site Site
	if err := visisql.NewSelectService(db).Get([]string{"id", "url", "image"}, TABLE_NAME, nil, predicates, nil, &site); err != nil {
		return nil, fmt.Errorf("get failed to fetch site: %w", err)
	}

	return &site, nil
}

func update(db *sql.DB, id interface{}) error {
	ts, err := visisql.NewTransactionService(db)
	if err != nil {
		return fmt.Errorf("update failed to init transaction: %w", err)
	}

	predicates := [][]*visisql.Predicate{{
		visisql.NewPredicate("id", visisql.OperatorEqual, []interface{}{id}),
	}}

	if err := ts.Update(TABLE_NAME, map[string]interface{}{"url": "url", "image": "image"}, predicates); err != nil {
		return fmt.Errorf("update failed to update data: %w", err)
	}

	if err := ts.Commit(); err != nil {
		return fmt.Errorf("update failed to commit transaction: %w", err)
	}

	return nil
}

func remove(db *sql.DB, id interface{}) error {
	ts, err := visisql.NewTransactionService(db)
	if err != nil {
		return fmt.Errorf("delete failed to init transaction: %w", err)
	}

	predicates := [][]*visisql.Predicate{{
		visisql.NewPredicate("id", visisql.OperatorEqual, []interface{}{id}),
	}}

	if err := ts.Delete(TABLE_NAME, predicates); err != nil {
		return fmt.Errorf("delete failed to delete row: %w", err)
	}

	if err := ts.Commit(); err != nil {
		return fmt.Errorf("delete failed to commit transaction: %w", err)
	}

	return nil
}

func insertMultiple(db *sql.DB) error {
	ts, err := visisql.NewTransactionService(db)
	if err != nil {
		return fmt.Errorf("insert multiple failed to init transaction: %w", err)
	}

	data := [][]interface{}{{
		"url 1", "image 1",
	}, {
		"url 2", "image 2",
	}, {
		"url 3", "image 3",
	}}

	_, err = ts.InsertMultiple(TABLE_NAME, []string{"url", "image"}, data, nil)
	if err != nil {
		return fmt.Errorf("insert multiple failed to insert data: %w", err)
	}

	if err := ts.Commit(); err != nil {
		return fmt.Errorf("insert multiple failed to commit transaction: %w", err)
	}

	return nil
}

func list(db *sql.DB) ([]*Site, error) {
	var sites []*Site
	if _, _, _, err := visisql.NewSelectService(db).Search([]string{"id", "url", "image"}, TABLE_NAME, nil, nil, nil, nil, nil, &sites); err != nil {
		return nil, fmt.Errorf("list failed to fetch sites: %w", err)
	}

	return sites, nil
}
