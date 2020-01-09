package visisql

import (
	"database/sql"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

type SqlService struct {
	db *sql.DB
}

func NewSqlService(db *sql.DB) *SqlService {
	return &SqlService{db: db}
}

func (ss *SqlService) Query(query string, args []interface{}, v interface{}) error {
	rows, err := ss.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	slice := reflect.ValueOf(v).Elem()
	for rows.Next() {
		data, err := ss.rowToMap(rows)
		if err != nil {
			return err
		}

		item := reflect.New(reflect.TypeOf(v).Elem().Elem().Elem())
		if err := ss.hydrateStruct(data, item.Interface()); err != nil {
			return err
		}

		slice.Set(reflect.Append(slice, item))
	}

	return nil
}

func (ss *SqlService) QueryRow(query string, args []interface{}, v interface{}) error {
	rows, err := ss.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}

		return sql.ErrNoRows
	}

	data, err := ss.rowToMap(rows)
	if err != nil {
		return err
	}

	return ss.hydrateStruct(data, v)
}

func (ss *SqlService) List(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, orderBy []*OrderBy, pagination *Pagination, v interface{}) (int64, int64, int64, error) {
	builderRs := sqlbuilder.PostgreSQL.NewSelectBuilder()

	// from
	builderRs.From(from)

	// joins
	for _, j := range joins {
		if j.option != "" {
			builderRs.JoinWithOption(j.option, j.table, j.on)
		} else {
			builderRs.Join(j.table, j.on)
		}
	}

	// predicates
	sPs, err := predicatesToStrings(predicates, &builderRs.Cond)
	if err != nil {
		return 0, 0, 0, err
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
		return 0, 0, 0, err
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

	if err = ss.QueryRow(queryC, argsC, &CountSql); err != nil {
		return 0, 0, 0, err
	}

	return CountSql.Count, CountSql.TotalCount, CountSql.PageCount, nil
}

func (ss *SqlService) Get(fields []string, from string, joins []*Join, predicates [][]*Predicate, groupBy []string, v interface{}) error {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()

	builder.Select(fields...)
	builder.From(from)

	for _, j := range joins {
		if j.option != "" {
			builder.JoinWithOption(j.option, j.table, j.on)
		} else {
			builder.Join(j.table, j.on)
		}
	}

	sPs, err := predicatesToStrings(predicates, &builder.Cond)
	if err != nil {
		return err
	}
	builder.Where(sPs...)

	builder.GroupBy(groupBy...)

	query, args := builder.Build()

	return ss.QueryRow(query, args, v)
}

func (ss *SqlService) Create(into string, values map[string]interface{}) (interface{}, *sql.Tx, error) {
	builder := sqlbuilder.PostgreSQL.NewInsertBuilder()

	builder.InsertInto(into)

	var fields []string
	var vals []interface{}
	for f, v := range values {
		fields = append(fields, f)
		vals = append(vals, v)
	}

	builder.Cols(fields...)
	builder.Values(vals...)

	query, args := builder.Build()

	tx, err := ss.db.Begin()
	if err != nil {
		return nil, nil, err
	}

	row := tx.QueryRow(fmt.Sprintf(`%s returning id`, query), args...)
	if row == nil {
		_ = tx.Rollback()
		return nil, nil, fmt.Errorf("failed to exec insert query")
	}

	var id interface{}
	if err := row.Scan(&id); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	return id, tx, nil
}

func (ss *SqlService) Update(table string, set map[string]interface{}, predicates [][]*Predicate) (*sql.Tx, error) {
	builder := sqlbuilder.PostgreSQL.NewUpdateBuilder()

	builder.Update(table)

	var str []string
	for f, v := range set {
		str = append(str, builder.Assign(f, v))
	}

	builder.Set(str...)

	sPs, err := predicatesToStrings(predicates, &builder.Cond)
	if err != nil {
		return nil, err
	}
	builder.Where(sPs...)

	query, args := builder.Build()

	tx, err := ss.db.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	return tx, nil
}

func (ss *SqlService) Delete(from string, predicates [][]*Predicate) error {
	builder := sqlbuilder.PostgreSQL.NewDeleteBuilder()

	builder.DeleteFrom(from)

	sPs, err := predicatesToStrings(predicates, &builder.Cond)
	if err != nil {
		return err
	}
	builder.Where(sPs...)

	query, args := builder.Build()

	_, err = ss.db.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (ss *SqlService) rowToMap(row *sql.Rows) (map[string]interface{}, error) {
	cols, err := row.Columns()
	if err != nil {
		return nil, err
	}

	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = &vals[i]
	}

	if err := row.Scan(vals...); err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for i, col := range cols {
		res[col] = vals[i]
	}

	return res, nil
}

func (ss *SqlService) hydrateStruct(data map[string]interface{}, v interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "sql", Result: v})
	if err != nil {
		return err
	}

	return decoder.Decode(data)
}
