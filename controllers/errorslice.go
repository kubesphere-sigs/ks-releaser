package controllers

import (
	"errors"
	"fmt"
)

// ErrorSlice is a collection of errors
type ErrorSlice []error

func (e ErrorSlice) append(err error) ErrorSlice {
	return append(e, err)
}

// String returns a readable string
func (e ErrorSlice) String() (result string) {
	for _, item := range e {
		result += fmt.Sprintf("%v", item)
	}
	return
}

// ToError turns to an error object
func (e ErrorSlice) ToError() error {
	msg := e.String()
	if msg == "" {
		return nil
	}
	return errors.New(msg)
}
