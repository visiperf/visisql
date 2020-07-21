package visisql

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
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
	type BuildReq struct {
		fields     []string
		from       string
		joins      []*Join
		predicates [][]*Predicate
		groupBy    []string
		orderBy    []*OrderBy
		pagination *Pagination
	}

	type BuildResp struct {
		query string
		args  []interface{}
		err   error
	}

	var tests = []struct {
		Req     *BuildReq
		Resp    *BuildResp
		Message string
	}{{
		Message: "bad predicates",
		Req: &BuildReq{
			fields: []string{"id", "email", "roles", "created_at"},
			from:   "user",
			joins:  nil,
			predicates: [][]*Predicate{{
				NewPredicate("id", OperatorEqual, []interface{}{1, 2}),
			}},
			groupBy:    nil,
			orderBy:    nil,
			pagination: nil,
		},
		Resp: &BuildResp{
			query: "",
			args:  nil,
			err:   &QueryError{errOperatorEqual},
		},
	}, {
		Message: "ok without order by and pagination",
		Req: &BuildReq{
			fields: []string{"id", "email", "roles", "created_at"},
			from:   "user",
			joins: []*Join{
				NewJoin(LeftJoin, "company", "user.company_id = company.id"),
			},
			predicates: [][]*Predicate{{
				NewPredicate("user.id", OperatorEqual, []interface{}{1}),
			}},
			groupBy:    []string{"user.id"},
			orderBy:    nil,
			pagination: nil,
		},
		Resp: &BuildResp{
			query: `SELECT id, email, roles, created_at FROM user LEFT JOIN company ON user.company_id = company.id WHERE ( user.id = $1 ) GROUP BY user.id`,
			args:  []interface{}{1},
			err:   nil,
		},
	}, {
		Message: "ok",
		Req: &BuildReq{
			fields: []string{"id", "email", "roles", "created_at"},
			from:   "user",
			joins: []*Join{
				NewJoin(LeftJoin, "company", "user.company_id = company.id"),
			},
			predicates: [][]*Predicate{{
				NewPredicate("user.id", OperatorEqual, []interface{}{1}),
			}},
			groupBy: []string{"user.id"},
			orderBy: []*OrderBy{
				NewOrderBy("user.id", OrderAsc),
			},
			pagination: NewPagination(1, 2),
		},
		Resp: &BuildResp{
			query: `SELECT id, email, roles, created_at FROM user LEFT JOIN company ON user.company_id = company.id WHERE ( user.id = $1 ) GROUP BY user.id ORDER BY user.id ASC LIMIT 2 OFFSET 1`,
			args:  []interface{}{1},
			err:   nil,
		},
	}}

	ss := NewSelectService(nil)

	for _, test := range tests {
		query, args, err := ss.Build(test.Req.fields, test.Req.from, test.Req.joins, test.Req.predicates, test.Req.groupBy, test.Req.orderBy, test.Req.pagination)

		assert.Equal(t, test.Resp.query, query, test.Message)
		assert.Equal(t, test.Resp.args, args, test.Message)

		if test.Resp.err != nil {
			assert.True(t, strings.Contains(err.Error(), test.Resp.err.Error()), test.Message)
		} else {
			assert.Nil(t, err, test.Message)
		}
	}
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
