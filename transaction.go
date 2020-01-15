package visisql

import (
	"database/sql"
	"errors"
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

// @todo: implement TransactionService.Insert()
func (ts *TransactionService) Insert(into string, values map[string]interface{}, returning interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
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

// @todo: implement TransactionService.Rollback()
func (ts *TransactionService) Rollback() error {
	return errors.New("not implemented")
}

// @todo: implement TransactionService.Commit()
func (ts *TransactionService) Commit() error {
	return errors.New("not implemented")
}
