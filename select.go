package visisql

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
)

type SelectService interface {
	Build(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []*OrderBy, pagination *Pagination) (string, []interface{}, error)
	Query(query string, args []interface{}, v interface{}) error
	QueryRow(query string, args []interface{}, v interface{}) error
	Search(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []*OrderBy, pagination *Pagination, v interface{}) (int64, int64, int64, error)
	Get(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, v interface{}) error
}

type selectService struct {
	db *sqlx.DB
}

func NewSelectService(db *sqlx.DB) SelectService {
	return &selectService{db: db}
}

func (ss *selectService) Build(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []*OrderBy, pagination *Pagination) (string, []interface{}, error) {
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
		return "", nil, err
	}
	builder.Where(sPs...)

	builder.GroupBy(groupBy...)

	var ob []string
	for _, o := range orderBy {
		ob = append(ob, o.toString())
	}
	builder.OrderBy(ob...)

	if pagination != nil {
		builder.Offset(pagination.Start)

		if pagination.Limit != 0 {
			builder.Limit(pagination.Limit)
		}
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

func (ss *selectService) Search(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []*OrderBy, pagination *Pagination, v interface{}) (int64, int64, int64, error) {
	return 0, 0, 0, nil
}

func (ss *selectService) Get(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, v interface{}) error {
	query, args, err := ss.Build(fields, from, joins, predicates, groupBy, nil, nil)
	if err != nil {
		return err
	}

	return ss.QueryRow(query, args, v)
}

func (ss *selectService) List(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []*OrderBy, pagination *Pagination, v interface{}) (int64, int64, int64, error) {
	builderRs := sqlbuilder.PostgreSQL.NewSelectBuilder()

	// from
	builderRs.From(from)

	// joins
	for _, j := range joins {
		if j.option == InnerJoin {
			builderRs.Join(j.table, j.on)
		} else {
			builderRs.JoinWithOption(sqlbuilder.JoinOption(j.option), j.table, j.on)
		}
	}

	// predicates
	sPs, err := predicatesToStrings(predicates, &builderRs.Cond)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("visisql list predicates: %w", err)
	}
	builderRs.Where(sPs...)

	// group by
	builderRs.GroupBy(groupBy...)

	// order by
	var ob []string
	for _, o := range orderBy {
		ob = append(ob, o.toString())
	}
	builderRs.OrderBy(ob...)

	// pagination
	if pagination != nil {
		builderRs.Offset(pagination.Start)

		if pagination.Limit != 0 {
			builderRs.Limit(pagination.Limit)
		}
	}

	builderRs.Select(fields...)
	queryRs, argsRs := builderRs.Build()

	if err := ss.Query(queryRs, argsRs, v); err != nil {
		return 0, 0, 0, fmt.Errorf("visisql list query: %w", err)
	}

	builderRs.Select("count(*) over () as total_count")

	builderC := sqlbuilder.PostgreSQL.NewSelectBuilder()

	// fields
	builderC.Select("count(*) as count", "total_count", "ceil(total_count::decimal / count(*))::integer as page_count")

	// from
	builderC.From(builderC.BuilderAs(builderRs, "results"))

	// group by
	builderC.GroupBy("total_count")

	queryC, argsC := builderC.Build()

	var CountSql = struct {
		Count      int64 `sql:"count"`
		TotalCount int64 `sql:"total_count"`
		PageCount  int64 `sql:"page_count"`
	}{}

	if err = ss.QueryRow(queryC, argsC, &CountSql); err != nil && err != sql.ErrNoRows {
		return 0, 0, 0, fmt.Errorf("visisql list count query: %w", err)
	}

	return CountSql.Count, CountSql.TotalCount, CountSql.PageCount, nil
}

func (ss *selectService) rowToMap(row *sql.Rows) (map[string]interface{}, error) {
	cols, err := row.Columns()
	if err != nil {
		return nil, fmt.Errorf("visisql row columns: %w", err)
	}

	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = &vals[i]
	}

	if err := row.Scan(vals...); err != nil {
		return nil, fmt.Errorf("visisql row scan: %w", err)
	}

	res := make(map[string]interface{})
	for i, col := range cols {
		res[col] = vals[i]
	}

	return res, nil
}

func (ss *selectService) hydrateStruct(data map[string]interface{}, v interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "sql", Result: v})
	if err != nil {
		return fmt.Errorf("visisql struct decoder: %w", err)
	}

	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("visisql struct hydration: %w", err)
	}

	return nil
}
