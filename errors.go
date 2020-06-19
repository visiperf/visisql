package visisql

type QueryError struct {
	err error
}

func (e *QueryError) Error() string {
	return e.err.Error()
}
