package visisql

import (
	"database/sql"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"reflect"
)

type SqlService struct {
	db *sql.DB
}

func NewSqlService(db *sql.DB) *SqlService {
	return &SqlService{db: db}
}

func (ss *SqlService) List(identifier string, fields []string, from string, joins []*Join, predicates []*Predicate, orderBy map[string]string, start int64, limit int64, v interface{}) error {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()

	var sCount = "1"
	if limit != 0 {
		sCount = fmt.Sprintf("ceil(count(*) over() / cast(%d as float))", limit)
	}
	sCount += " as page_count"

	builder.Select(append(fields, sCount)...)
	builder.From(from)

	for _, j := range joins {
		if j.option != "" {
			builder.JoinWithOption(j.option, j.table, j.on)
		} else {
			builder.Join(j.table, j.on)
		}
	}

	for _, p := range predicates {
		if p.IsOperator(OperatorIn) {
			builder.Where(builder.In(p.Field, p.Values...))
		}
		if p.IsOperator(OperatorEqual) {
			if len(p.Values) != 1 {
				return fmt.Errorf(`predicate must have only one value when operator is equal`)
			}
			builder.Where(builder.Equal(p.Field, p.Values[0]))
		}
		if p.IsOperator(OperatorLike) {
			if len(p.Values) != 1 {
				return fmt.Errorf(`predicate must have only one value when operator is like`)
			}
			builder.Where(builder.Like(p.Field, p.Values[0]))
		}
	}

	builder.GroupBy(identifier)

	var ob []string
	for f, v := range orderBy {
		ob = append(ob, fmt.Sprintf("%s %s", f, v))
	}
	builder.OrderBy(ob...)

	if start != 0 {
		builder.Offset(int(start))
	}

	if limit != 0 {
		builder.Limit(int(limit))
	}

	query, args := builder.Build()

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

func (ss *SqlService) Get(identifier string, fields []string, from string, joins []*Join, predicates []*Predicate, v interface{}) error {
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

	for _, p := range predicates {
		if p.IsOperator(OperatorIn) {
			builder.Where(builder.In(p.Field, p.Values...))
		}
		if p.IsOperator(OperatorEqual) {
			if len(p.Values) != 1 {
				return fmt.Errorf(`predicate must have only one value when operator is equal`)
			}
			builder.Where(builder.Equal(p.Field, p.Values[0]))
		}
		if p.IsOperator(OperatorLike) {
			if len(p.Values) != 1 {
				return fmt.Errorf(`predicate must have only one value when operator is like`)
			}
			builder.Where(builder.Like(p.Field, p.Values[0]))
		}
	}

	builder.GroupBy(identifier)

	query, args := builder.Build()

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

func (ss *SqlService) Create(into string, values map[string]interface{}) (interface{}, error) {
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

	row := ss.db.QueryRow(fmt.Sprintf(`%s returning id`, query), args...)
	if row == nil {
		return nil, fmt.Errorf("row is nil")
	}

	var id interface{}
	if err := row.Scan(&id); err != nil {
		return nil, err
	}

	return id, nil
}

func (ss *SqlService) Update(tableName string, set map[string]interface{}, predicates []*Predicate) error {
	builder := sqlbuilder.PostgreSQL.NewUpdateBuilder()

	builder.Update(tableName)

	var str []string
	for f, v := range set {
		str = append(str, builder.Assign(f, v))
	}

	builder.Set(str...)

	for _, p := range predicates {
		if p.IsOperator(OperatorIn) {
			builder.Where(builder.In(p.Field, p.Values...))
		}
		if p.IsOperator(OperatorEqual) {
			if len(p.Values) != 1 {
				return fmt.Errorf(`predicate must have only one value when operator is equal`)
			}
			builder.Where(builder.Equal(p.Field, p.Values[0]))
		}
		if p.IsOperator(OperatorLike) {
			if len(p.Values) != 1 {
				return fmt.Errorf(`predicate must have only one value when operator is like`)
			}
			builder.Where(builder.Like(p.Field, p.Values[0]))
		}
	}

	query, args := builder.Build()

	_, err := ss.db.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (ss *SqlService) Delete(from string, predicates []*Predicate) error {
	builder := sqlbuilder.PostgreSQL.NewDeleteBuilder()

	builder.DeleteFrom(from)

	for _, p := range predicates {
		if p.IsOperator(OperatorIn) {
			builder.Where(builder.In(p.Field, p.Values...))
		}
		if p.IsOperator(OperatorEqual) {
			if len(p.Values) != 1 {
				return fmt.Errorf(`predicate must have only one value when operator is equal`)
			}
			builder.Where(builder.Equal(p.Field, p.Values[0]))
		}
		if p.IsOperator(OperatorLike) {
			if len(p.Values) != 1 {
				return fmt.Errorf(`predicate must have only one value when operator is like`)
			}
			builder.Where(builder.Like(p.Field, p.Values[0]))
		}
	}

	query, args := builder.Build()

	_, err := ss.db.Exec(query, args...)
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
	fields := reflect.ValueOf(v).Elem()

	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)

		if field.Kind() == reflect.Struct {
			if err := ss.hydrateStruct(data, field.Addr().Interface()); err != nil {
				return err
			}
		} else {
			if !field.CanSet() {
				return fmt.Errorf("cannot set field %d", i)
			}

			tag := fields.Type().Field(i).Tag.Get("sql")

			if _, ok := data[tag]; ok {
				if val := reflect.ValueOf(data[tag]); val.IsValid() {
					if field.Type() != val.Type() {
						return fmt.Errorf("provided value type didn't match obj field type, expected %s - actual %s", field.Type().String(), val.Type().String())
					}

					field.Set(val)
				}
			}
		}
	}

	return nil
}
