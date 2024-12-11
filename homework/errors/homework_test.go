package main

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type MultiError struct {
	errs []error
}

func (e *MultiError) Error() string {
	if e == nil || len(e.errs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(strconv.Itoa(len(e.errs)))
	sb.WriteString(" errors occurred:\n")
	for _, err := range e.errs {
		sb.WriteString("\t* ")
		sb.WriteString(err.Error())
	}
	sb.WriteString("\n")
	return sb.String()
}

func Append(err error, errs ...error) (m *MultiError) {
	if err == nil {
		m = &MultiError{errs: errs}
		return
	}

	var ok bool
	m, ok = err.(*MultiError)
	if !ok {
		panic("invalid error type")
	}
	m.errs = append(m.errs, errs...)
	return
}

func (e *MultiError) Is(target error) bool {
	for _, err := range e.errs {
		if err == target {
			return true
		}
	}
	return false
}

func (e *MultiError) As(target any) bool {
	for _, err := range e.errs {
		if reflect.TypeOf(err) == reflect.TypeOf(target).Elem() {
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(err))
			return true
		}
	}
	return false
}

func (e *MultiError) Unwrap() error {
	if len(e.errs) <= 1 {
		return nil
	}
	return &MultiError{errs: e.errs[:len(e.errs)-1]}
}

func TestMultiError(t *testing.T) {
	var err error
	err = Append(err, errors.New("error 1"))
	err = Append(err, errors.New("error 2"))

	expectedMessage := "2 errors occurred:\n\t* error 1\t* error 2\n"
	assert.EqualError(t, err, expectedMessage)
}

func TestMultiError_Is(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	var err error
	err = Append(err, err1)
	err = Append(err, err2)

	assert.False(t, errors.Is(err, err3))
	assert.True(t, errors.Is(err, err2))
	assert.True(t, errors.Is(err, err1))

	assert.True(t, errors.Is(err, err1))
	assert.True(t, errors.Is(err, err2))
	assert.False(t, errors.Is(err, err3))
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }

func TestMultiError_As(t *testing.T) {
	var err error
	err = Append(err, errors.New("error 1"))
	err = Append(err, &testError{msg: "error 2"})

	var te *testError
	assert.True(t, errors.As(err, &te))
	assert.True(t, te.msg == "error 2")
}

func TestMultiError_Unwrap(t *testing.T) {
	var err error
	err = Append(err, errors.New("error 1"))
	err = Append(err, errors.New("error 2"))

	exp := "2 errors occurred:\n\t* error 1\t* error 2\n"
	assert.EqualError(t, err, exp)

	err = errors.Unwrap(err)
	exp = "1 errors occurred:\n\t* error 1\n"
	assert.EqualError(t, err, exp)

	err = errors.Unwrap(err)
	assert.Nil(t, err)
}
