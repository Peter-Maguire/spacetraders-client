package http

type HttpResponse[T any] struct {
	Data          *T             `json:"data"`
	PaginatedMeta *PaginatedMeta `json:"meta"`
	Error         *HttpError     `json:"error"`
}

type PaginatedMeta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type HttpError struct {
	Message       string
	Code          ErrorCode
	Data          map[string]any
	InternalError error
}

func (e *HttpError) Error() string {
	return e.Message
}

func InternalError(err error) *HttpError {
	return &HttpError{InternalError: err}
}
