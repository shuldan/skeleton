package api

import "github.com/shuldan/errors"

var apiCode = errors.WithPrefix("TASK_API")

var (
	ErrInvalidTaskID = apiCode("INVALID_TASK_ID").
		Kind(errors.Validation).
		New("invalid task id")
)
