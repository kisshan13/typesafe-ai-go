package typesafe

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type APIError struct {
	StatusCode int
	Status     string
	Body       string
	RetryAfter string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("typesafe: API returned %s", e.Status)
	}
	return fmt.Sprintf("typesafe: API returned %s: %s", e.Status, e.Body)
}

func (e *APIError) Unwrap() error {
	return nil
}

func (e *APIError) RetryAfterDuration() time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(e.RetryAfter))
	if err != nil || seconds < 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func validateRequest(req EvaluationRequest) error {
	if req.Questions == nil || len(req.Questions) == 0 {
		return errors.New("typesafe: at least one question is required")
	}
	for id, question := range req.Questions {
		if strings.TrimSpace(id) == "" {
			return errors.New("typesafe: question id cannot be empty")
		}
		if question == nil {
			return fmt.Errorf("typesafe: question %q cannot be nil", id)
		}
	}
	return nil
}
