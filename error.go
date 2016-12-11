package main

import (
	"github.com/scgolang/nsm"
)

// Error represents the data contained in an error response from a non session manager.
type Error struct {
	nsmErr  nsm.Error
	Address string
}

func (e Error) Code() nsm.Code {
	return e.nsmErr.Code()
}

func (e Error) Error() string {
	return e.nsmErr.Error()
}

// NewError creates a new error.
func NewError(nsmErr nsm.Error, addr string) Error {
	return Error{nsmErr: nsmErr, Address: addr}
}
