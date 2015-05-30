package form

type TagError struct {
	err error
}

func (t TagError) Error() string {
	return t.err.Error()
}
