package visisql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
)

type TransactionService struct {
	tx *sql.Tx
}

func NewTransactionService(db *sql.DB) (*TransactionService, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
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
				return nil, rErr
			}
			return nil, err
		}
	} else {
		if _, err := ts.tx.Exec(query, args...); err != nil {
			if rErr := ts.tx.Rollback(); rErr != nil {
				return nil, rErr
			}
			return nil, err
		}
	}

	return resp, nil
}

// @todo: implement TransactionService.InsertMultiple()
func (ts *TransactionService) InsertMultiple(into string, values []map[string]interface{}, returning interface{}) ([]interface{}, error) {
	return nil, errors.New("not implemented")
}

// @todo: implement TransactionService.Update()
func (ts *TransactionService) Update(table string, set map[string]interface{}, predicates [][]*Predicate) error {
	return errors.New("not implemented")
}

// @todo: implement TransactionService.Delete()
func (ts *TransactionService) Delete(from string, predicates [][]*Predicate) error {
	return errors.New("not implemented")
}

func (ts *TransactionService) Rollback() error {
	return ts.tx.Rollback()
}

func (ts *TransactionService) Commit() error {
	return ts.tx.Commit()
}
