package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/visiperf/visisql/v2"
)

type Phones []string

func (ps *Phones) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &ps)
}

type Company struct {
	ID        int64      `db:"id"`
	Name      string     `db:"name"`
	Phones    *Phones    `db:"phones"`
	CreatedAt *time.Time `db:"created_at"`
}

var schema = struct {
	tableName string
	fields    []string
	create    string
	mocks     string
	drop      string
}{
	tableName: `company`,
	fields:    []string{`id`, `name`, `phones`, `created_at`},
	create: `
		create table company (
			id 			serial 		primary key,
			name 		varchar 	unique not null,
			phones 		jsonb,
			created_at 	timestamptz not null default now()
		);
	`,
	mocks: `
		insert into company (name, phones) values 
			('Google', '["01.02.03.04.05", "02.03.04.05.06"]'),
			('Apple', '["03.04.05.06.07"]'),
			('Facebook', '[]'),
			('Amazon', null);
	`,
	drop: `
		drop table company;
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

func main() {
	db, err := openDatabase()
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	q, args, err := build(db)
	if err != nil {
		log.Fatalf("failed to build: %v", err)
	}
	fmt.Println("query: ", q)
	fmt.Println("args: ", args)

	companies, err := query(db)
	if err != nil {
		log.Fatalf("failed to query: %v", err)
	}
	for _, c := range companies {
		fmt.Println(c, c.Phones)
	}

	company, err := queryRow(db)
	if err != nil {
		log.Fatalf("failed to query row: %v", err)
	}
	fmt.Println(company, company.Phones)

	company, err = get(db)
	if err != nil {
		log.Fatalf("failed to get: %v", err)
	}
	fmt.Println(company, company.Phones)

	companies, c, tc, pc, err := search(db)
	if err != nil {
		log.Fatalf("failed to query: %v", err)
	}
	for _, c := range companies {
		fmt.Println(c, c.Phones)
	}
	fmt.Println(c, tc, pc)
}

func build(db *sqlx.DB) (string, []interface{}, error) {
	return visisql.NewSelectService(db).Build(schema.fields, schema.tableName, nil, [][]*visisql.Predicate{{
		visisql.NewPredicate("id", visisql.OperatorEqual, []interface{}{1}),
	}}, nil, nil, nil)
}

func query(db *sqlx.DB) ([]*Company, error) {
	defer func() {
		db.Exec(schema.drop)
	}()

	if _, err := db.Exec(schema.create); err != nil {
		return nil, err
	}

	if _, err := db.Exec(schema.mocks); err != nil {
		return nil, err
	}

	var companies []*Company
	if err := visisql.NewSelectService(db).Query(`SELECT * FROM company`, nil, &companies); err != nil {
		return nil, err
	}

	return companies, nil
}

func queryRow(db *sqlx.DB) (*Company, error) {
	defer func() {
		db.Exec(schema.drop)
	}()

	if _, err := db.Exec(schema.create); err != nil {
		return nil, err
	}

	if _, err := db.Exec(schema.mocks); err != nil {
		return nil, err
	}

	var company Company
	if err := visisql.NewSelectService(db).QueryRow(`SELECT * FROM company WHERE id = $1`, []interface{}{1}, &company); err != nil {
		return nil, err
	}

	return &company, nil
}

func get(db *sqlx.DB) (*Company, error) {
	defer func() {
		db.Exec(schema.drop)
	}()

	if _, err := db.Exec(schema.create); err != nil {
		return nil, err
	}

	if _, err := db.Exec(schema.mocks); err != nil {
		return nil, err
	}

	var company Company
	if err := visisql.NewSelectService(db).Get(schema.fields, schema.tableName, nil, [][]*visisql.Predicate{{
		visisql.NewPredicate("id", visisql.OperatorEqual, []interface{}{2}),
	}}, nil, &company); err != nil {
		return nil, err
	}

	return &company, nil
}

func search(db *sqlx.DB) ([]*Company, int64, int64, int64, error) {
	defer func() {
		db.Exec(schema.drop)
	}()

	if _, err := db.Exec(schema.create); err != nil {
		return nil, 0, 0, 0, err
	}

	if _, err := db.Exec(schema.mocks); err != nil {
		return nil, 0, 0, 0, err
	}

	var companies []*Company

	c, tc, pc, err := visisql.NewSelectService(db).Search(schema.fields, schema.tableName, nil, nil, nil, []*visisql.OrderBy{
		visisql.NewOrderBy("id", visisql.OrderAsc),
	}, nil, &companies)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	return companies, c, tc, pc, nil
}
