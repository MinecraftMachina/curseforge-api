package util

import (
	"fmt"
	"github.com/pkg/errors"
)

var (
	ErrNon200StatusCode = errors.New("non-200 status code")
)

func CreateNon200Error(code int, body []byte) error {
	return errors.WithMessage(errors.WithStack(ErrNon200StatusCode),
		fmt.Sprintf("code: %d, body: %s", code, string(body)))
}
