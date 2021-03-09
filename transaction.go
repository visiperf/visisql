package visisql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
)

type TransactionService interface {
	InsertOnConflictUpdate(into string, conflictOn []string, values map[string]interface{}, returning interface{}) (interface{}, error)
	Insert(into string, values map[string]interface{}, returning interface{}) (interface{}, error)
	InsertMultiple(into string, fields []string, values [][]interface{}, returning interface{}) ([]interface{}, error)
	Update(table string, set map[string]interface{}, predicates [][]*Predicate) error
	Delete(from string, predicates [][]*Predicate) error
	Rollback() error
	Commit() error
}

type transactionService struct {
	tx *sql.Tx
}

func NewTransactionService(db *sqlx.DB) (TransactionService, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	return &transactionService{tx: tx}, nil
}

func (ts *transactionService) InsertOnConflictUpdate(into string, conflictOn []string, values map[string]interface{}, returning interface{}) (interface{}, error) {
	columns := extractMapKeys(values)

	insertBuilder := sqlbuilder.PostgreSQL.NewInsertBuilder()

	insertBuilder.
		InsertInto(into).
		Cols(columns...).
		Values(extractMapValues(values, columns)...)

	insertQuery, insertArgs := insertBuilder.Build()

	updateBuilder := sqlbuilder.PostgreSQL.NewUpdateBuilder()

	updateBuilder.Set(assignMap(values, columns, updateBuilder)...)

	updateQuery, updateArgs := updateBuilder.Build()

	if len(insertArgs) != len(updateArgs) {
		return nil, fmt.Errorf("something went wrong with assignements, insert has %d args but update has %d", len(insertArgs), len(updateArgs))
	}

	query := fmt.Sprintf("%s on conflict (%s) do %s", insertQuery, strings.Join(conflictOn, ","), updateQuery)

	var resp interface{}
	if returning != nil {
		row := ts.tx.QueryRow(fmt.Sprintf("%s returning %s", query, returning), insertArgs...)
		if err := row.Scan(&resp); err != nil {
			if rErr := ts.tx.Rollback(); rErr != nil {
				return nil, fmt.Errorf("visisql rollback: %w", rErr)
			}
			return nil, fmt.Errorf("visisql query: %w", &QueryError{err})
		}
	} else {
		if _, err := ts.tx.Exec(query, insertArgs...); err != nil {
			if rErr := ts.tx.Rollback(); rErr != nil {
				return nil, fmt.Errorf("visisql rollback: %w", rErr)
			}
			return nil, fmt.Errorf("visisql query: %w", &QueryError{err})
		}
	}

	return resp, nil
}

func (ts *transactionService) Insert(into string, values map[string]interface{}, returning interface{}) (interface{}, error) {
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

	var resp interface{}
	if returning != nil {
		row := ts.tx.QueryRow(fmt.Sprintf("%s returning %s", query, returning), args...)
		if err := row.Scan(&resp); err != nil {
			if rErr := ts.tx.Rollback(); rErr != nil {
				return nil, fmt.Errorf("visisql rollback: %w", rErr)
			}
			return nil, fmt.Errorf("visisql query: %w", &QueryError{err})
		}
	} else {
		if _, err := ts.tx.Exec(query, args...); err != nil {
			if rErr := ts.tx.Rollback(); rErr != nil {
				return nil, fmt.Errorf("visisql rollback: %w", rErr)
			}
			return nil, fmt.Errorf("visisql query: %w", &QueryError{err})
		}
	}

	return resp, nil
}

func (ts *transactionService) InsertMultiple(into string, fields []string, values [][]interface{}, returning interface{}) ([]interface{}, error) {
	builder := sqlbuilder.PostgreSQL.NewInsertBuilder()

	builder.InsertInto(into)

	builder.Cols(fields...)
	builder.Values(values[0]...)

	query, _ := builder.Build()
	if returning != nil {
		query = fmt.Sprintf("%s returning %s", query, returning)
	}

	stmt, err := ts.tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("visisql statement prepare: %w", &QueryError{err})
	}

	var resps []interface{}
	for _, args := range values {
		if returning != nil {
			var resp interface{}

			row := stmt.QueryRow(args...)
			if err := row.Scan(&resp); err != nil {
				if rErr := ts.tx.Rollback(); rErr != nil {
					return nil, fmt.Errorf("visisql rollback: %w", rErr)
				}
				return nil, fmt.Errorf("visisql query: %w", &QueryError{err})
			}

			resps = append(resps, resp)
		} else {
			if _, err := stmt.Exec(args...); err != nil {
				if rErr := ts.tx.Rollback(); rErr != nil {
					return nil, fmt.Errorf("visisql rollback: %w", rErr)
				}
				return nil, fmt.Errorf("visisql query: %w", &QueryError{err})
			}
		}
	}

	return resps, nil
}

func (ts *transactionService) Update(table string, set map[string]interface{}, predicates [][]*Predicate) error {
	builder := sqlbuilder.PostgreSQL.NewUpdateBuilder()

	builder.Update(table)

	var str []string
	for f, v := range set {
		str = append(str, builder.Assign(f, v))
	}

	builder.Set(str...)

	sPs, err := predicatesToStrings(predicates, &builder.Cond)
	if err != nil {
		return err
	}
	builder.Where(sPs...)

	query, args := builder.Build()

	if _, err := ts.tx.Exec(query, args...); err != nil {
		if rErr := ts.tx.Rollback(); rErr != nil {
			return fmt.Errorf("visisql rollback: %w", rErr)
		}
		return fmt.Errorf("visisql query: %w", &QueryError{err})
	}

	return nil
}

func (ts *transactionService) Delete(from string, predicates [][]*Predicate) error {
	builder := sqlbuilder.PostgreSQL.NewDeleteBuilder()

	builder.DeleteFrom(from)

	sPs, err := predicatesToStrings(predicates, &builder.Cond)
	if err != nil {
		return err
	}
	builder.Where(sPs...)

	query, args := builder.Build()

	if _, err := ts.tx.Exec(query, args...); err != nil {
		if rErr := ts.tx.Rollback(); rErr != nil {
			return fmt.Errorf("visisql rollback: %w", rErr)
		}
		return fmt.Errorf("visisql query: %w", &QueryError{err})
	}

	return nil
}

func (ts *transactionService) Rollback() error {
	return ts.tx.Rollback()
}

func (ts *transactionService) Commit() error {
	return ts.tx.Commit()
}
