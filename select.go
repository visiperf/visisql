package visisql

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
)

type SelectService interface {
	Build(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []OrderBy, pagination *Pagination) (string, []interface{}, error)
	Query(query string, args []interface{}, v interface{}) error
	QueryRow(query string, args []interface{}, v interface{}) error
	Search(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []OrderBy, pagination *Pagination, v interface{}) (int64, int64, int64, error)
	Get(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, v interface{}) error
}

type selectService struct {
	db *sqlx.DB
}

func NewSelectService(db *sqlx.DB) SelectService {
	return &selectService{db: db}
}

func (ss *selectService) Build(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []OrderBy, pagination *Pagination) (string, []interface{}, error) {
	builder, err := ss.newBuilder(fields, from, joins, predicates, groupBy, orderBy, pagination)
	if err != nil {
		return "", nil, err
	}

	query, args := builder.Build()

	return query, args, nil
}

func (ss *selectService) Query(query string, args []interface{}, v interface{}) error {
	rows, err := ss.db.Queryx(query, args...)
	if err != nil {
		return fmt.Errorf("visisql query execution: %w", &QueryError{err})
	}
	defer rows.Close()

	slice := reflect.ValueOf(v).Elem()
	for rows.Next() {
		item := reflect.New(reflect.TypeOf(v).Elem().Elem().Elem())

		if err := rows.StructScan(item.Interface()); err != nil {
			return fmt.Errorf("visisql struct scan: %w", &ScanError{err})
		}

		slice.Set(reflect.Append(slice, item))
	}

	return nil
}

func (ss *selectService) QueryRow(query string, args []interface{}, v interface{}) error {
	rows, err := ss.db.Queryx(query, args...)
	if err != nil {
		return fmt.Errorf("visisql query execution: %w", &QueryError{err})
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return fmt.Errorf("visisql query execution: %w", &QueryError{err})
		}

		return fmt.Errorf("visisql row parsing: %w", sql.ErrNoRows)
	}

	if err := rows.StructScan(v); err != nil {
		return fmt.Errorf("visisql struct scan: %w", &ScanError{err})
	}

	return nil
}

func (ss *selectService) Search(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []OrderBy, pagination *Pagination, v interface{}) (int64, int64, int64, error) {
	builderRs, err := ss.newBuilder(fields, from, joins, predicates, groupBy, orderBy, pagination)
	if err != nil {
		return 0, 0, 0, err
	}

	queryRs, argsRs := builderRs.Build()

	if err := ss.Query(queryRs, argsRs, v); err != nil {
		return 0, 0, 0, fmt.Errorf("visisql records: %w", err)
	}

	if reflect.ValueOf(v).Elem().IsNil() {
		return 0, 0, 0, nil
	}

	builderRs.Select("count(*) over () as total_count")

	builderC := sqlbuilder.PostgreSQL.NewSelectBuilder()

	builderC.Select("count(*) as count", "total_count", "ceil(total_count::decimal / count(*))::integer as page_count")
	builderC.From(builderC.BuilderAs(builderRs, "results"))
	builderC.GroupBy("total_count")

	queryC, argsC := builderC.Build()

	var CountSql = struct {
		Count      int64 `db:"count"`
		TotalCount int64 `db:"total_count"`
		PageCount  int64 `db:"page_count"`
	}{}

	if err = ss.QueryRow(queryC, argsC, &CountSql); err != nil && err != sql.ErrNoRows {
		return 0, 0, 0, fmt.Errorf("visisql count: %w", err)
	}

	return CountSql.Count, CountSql.TotalCount, CountSql.PageCount, nil
}

func (ss *selectService) Get(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, v interface{}) error {
	query, args, err := ss.Build(fields, from, joins, predicates, groupBy, nil, nil)
	if err != nil {
		return err
	}

	return ss.QueryRow(query, args, v)
}

func (ss *selectService) newBuilder(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []OrderBy, pagination *Pagination) (*sqlbuilder.SelectBuilder, error) {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()

	builder.Select(fields...)
	builder.From(from)

	for _, j := range joins {
		if j.option == InnerJoin {
			builder.Join(j.table, j.on)
		} else {
			builder.JoinWithOption(sqlbuilder.JoinOption(j.option), j.table, j.on)
		}
	}

	sPs, err := predicatesToStrings(predicates, &builder.Cond)
	if err != nil {
		return nil, err
	}
	builder.Where(sPs...)

	builder.GroupBy(groupBy...)

	var ob []string
	for _, o := range orderBy {
		ob = append(ob, fmt.Sprintf("%s %s", o.GetField(), o.GetOrder()))
	}
	builder.OrderBy(ob...)

	if pagination != nil {
		builder.Offset(pagination.Start)

		if pagination.Limit != 0 {
			builder.Limit(pagination.Limit)
		}
	}

	return builder, nil
}
