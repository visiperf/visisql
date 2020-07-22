package visisql

type QueryError struct {
	err error
}

func (e *QueryError) Error() string {
	return e.err.Error()
}

type ScanError struct {
	err error
}

func (e *ScanError) Error() string {
	return e.err.Error()
}
