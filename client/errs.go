package client

import (
	"errors"
	"fmt"
)

const errorPrefix = "cmdstream client: "

// ErrClosed happens when the Client is closed while connected to the server.
var ErrClosed = errors.New("closed")

func wrapErr(err error) error {
	return fmt.Errorf(errorPrefix+"%w", err)
}
