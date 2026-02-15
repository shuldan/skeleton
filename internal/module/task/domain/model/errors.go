package model

import "github.com/shuldan/errors"

var taskCode = errors.WithPrefix("TASK")

var (
	ErrTaskNotFound = taskCode("NOT_FOUND").
			Kind(errors.NotFound).
			New("task {{.id}} not found")

	ErrInvalidStatusTransition = taskCode("INVALID_STATUS_TRANSITION").
					Kind(errors.DomainRule).
					New("cannot transition from {{.from}} to {{.to}}")

	ErrTitleRequired = taskCode("TITLE_REQUIRED").
				Kind(errors.Validation).
				New("task title is required")

	ErrConcurrentModification = taskCode("CONCURRENT_MODIFICATION").
					Kind(errors.Conflict).
					New("task {{.id}} was modified concurrently")
)
