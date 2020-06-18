package visisql

import (
	"database/sql"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
)

type TransactionService struct {
	tx *sql.Tx
}

func NewTransactionService(db *sql.DB) (*TransactionService, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("visisql transaction begin: %w", err)
	}

	return &TransactionService{tx: tx}, nil
}

func (ts *TransactionService) Insert(into string, values map[string]interface{}, returning interface{}) (interface{}, error) {
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
				return nil, fmt.Errorf("visisql insert rollback: %w", rErr)
			}
			return nil, fmt.Errorf("visisql insert query: %w", &QueryError{err})
		}
	} else {
		if _, err := ts.tx.Exec(query, args...); err != nil {
			if rErr := ts.tx.Rollback(); rErr != nil {
				return nil, fmt.Errorf("visisql insert rollback: %w", rErr)
			}
			return nil, fmt.Errorf("visisql insert query: %w", &QueryError{err})
		}
	}

	return resp, nil
}

func (ts *TransactionService) InsertMultiple(into string, fields []string, values [][]interface{}, returning interface{}) ([]interface{}, error) {
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
		return nil, fmt.Errorf("visisql insert multiple statement preparation: %w", err)
	}

	var resps []interface{}
	for _, args := range values {
		if returning != nil {
			var resp interface{}

			row := stmt.QueryRow(args...)
			if err := row.Scan(&resp); err != nil {
				if rErr := ts.tx.Rollback(); rErr != nil {
					return nil, fmt.Errorf("visisql insert multiple rollback: %w", rErr)
				}
				return nil, fmt.Errorf("visisql insert multiple query: %w", &QueryError{err})
			}

			resps = append(resps, resp)
		} else {
			if _, err := stmt.Exec(args...); err != nil {
				if rErr := ts.tx.Rollback(); rErr != nil {
					return nil, fmt.Errorf("visisql insert multiple rollback: %w", rErr)
				}
				return nil, fmt.Errorf("visisql insert multiple query: %w", &QueryError{err})
			}
		}
	}

	return resps, nil
}

func (ts *TransactionService) Update(table string, set map[string]interface{}, predicates [][]*Predicate) error {
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
			return rErr
		}
		return err
	}

	return nil
}

func (ts *TransactionService) Delete(from string, predicates [][]*Predicate) error {
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
			return rErr
		}
		return err
	}

	return nil
}

func (ts *TransactionService) Rollback() error {
	if err := ts.tx.Rollback(); err != nil {
		return fmt.Errorf("visisql rollback: %w", err)
	}

	return nil
}

func (ts *TransactionService) Commit() error {
	if err := ts.tx.Commit(); err != nil {
		return fmt.Errorf("visisql commit: %w", err)
	}

	return nil
}
