package visisql

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type User struct {
	ID        string     `db:"id"`
	Email     string     `db:"email"`
	Roles     []string   `db:"roles"`
	CreatedAt *time.Time `db:"created_at"`
}

type Schema struct {
	create string
	drop   string
}

var userSchema = Schema{
	create: `
		create table "user" (
			id 			serial 		primary key,
			email 		varchar 	unique not null,
			roles 		jsonb,
			created_at 	timestamptz not null default now()
		);
	`,
	drop: `
		drop table "user";
	`,
}

func openDatabase() (*sqlx.DB, error) {
	pghost := os.Getenv("PG_HOST")
	pgport := os.Getenv("PG_PORT")
	pguser := os.Getenv("PG_USER")
	pgpwd := os.Getenv("PG_PWD")
	pgdbname := os.Getenv("PG_DB_NAME")
	pgoptions := os.Getenv("PG_OPTIONS")

	return sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s %s", pghost, pgport, pguser, pgpwd, pgdbname, pgoptions))
}

func wrapper(schema Schema, t *testing.T, test func(db *sqlx.DB)) {
	db, err := openDatabase()
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if _, err := db.Exec(schema.create); err != nil {
		t.Errorf("failed to create schema: %v", err)
	}

	test(db)

	if _, err := db.Exec(schema.drop); err != nil {
		t.Errorf("failed to drop schema: %v", err)
	}
}

func TestBuild(t *testing.T) {

}

func TestQuery(t *testing.T) {
	wrapper(userSchema, t, func(db *sqlx.DB) {})
}

func TestQueryRow(t *testing.T) {
	wrapper(userSchema, t, func(db *sqlx.DB) {})
}

func TestSearch(t *testing.T) {
	wrapper(userSchema, t, func(db *sqlx.DB) {})
}

func TestGet(t *testing.T) {
	wrapper(userSchema, t, func(db *sqlx.DB) {})
}
