// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: primes/v1/primes.proto

package pb

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on RandomRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *RandomRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on RandomRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in RandomRequestMultiError, or
// nil if none found.
func (m *RandomRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *RandomRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetMin() < 2 {
		err := RandomRequestValidationError{
			field:  "Min",
			reason: "value must be greater than or equal to 2",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if m.GetMax() > 9999999999 {
		err := RandomRequestValidationError{
			field:  "Max",
			reason: "value must be less than or equal to 9999999999",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return RandomRequestMultiError(errors)
	}

	return nil
}

// RandomRequestMultiError is an error wrapping multiple validation errors
// returned by RandomRequest.ValidateAll() if the designated constraints
// aren't met.
type RandomRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m RandomRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m RandomRequestMultiError) AllErrors() []error { return m }

// RandomRequestValidationError is the validation error returned by
// RandomRequest.Validate if the designated constraints aren't met.
type RandomRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RandomRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RandomRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RandomRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RandomRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RandomRequestValidationError) ErrorName() string { return "RandomRequestValidationError" }

// Error satisfies the builtin error interface
func (e RandomRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRandomRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RandomRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RandomRequestValidationError{}

// Validate checks the field values on RandomResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *RandomResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on RandomResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in RandomResponseMultiError,
// or nil if none found.
func (m *RandomResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *RandomResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Prime

	if len(errors) > 0 {
		return RandomResponseMultiError(errors)
	}

	return nil
}

// RandomResponseMultiError is an error wrapping multiple validation errors
// returned by RandomResponse.ValidateAll() if the designated constraints
// aren't met.
type RandomResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m RandomResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m RandomResponseMultiError) AllErrors() []error { return m }

// RandomResponseValidationError is the validation error returned by
// RandomResponse.Validate if the designated constraints aren't met.
type RandomResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RandomResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RandomResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RandomResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RandomResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RandomResponseValidationError) ErrorName() string { return "RandomResponseValidationError" }

// Error satisfies the builtin error interface
func (e RandomResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRandomResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RandomResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RandomResponseValidationError{}

// Validate checks the field values on ListRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ListRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ListRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ListRequestMultiError, or
// nil if none found.
func (m *ListRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ListRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetMin() < 2 {
		err := ListRequestValidationError{
			field:  "Min",
			reason: "value must be greater than or equal to 2",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if m.GetMax() > 9999999999 {
		err := ListRequestValidationError{
			field:  "Max",
			reason: "value must be less than or equal to 9999999999",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if val := m.GetMaxResults(); val < 0 || val > 5000 {
		err := ListRequestValidationError{
			field:  "MaxResults",
			reason: "value must be inside range [0, 5000]",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return ListRequestMultiError(errors)
	}

	return nil
}

// ListRequestMultiError is an error wrapping multiple validation errors
// returned by ListRequest.ValidateAll() if the designated constraints aren't met.
type ListRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ListRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ListRequestMultiError) AllErrors() []error { return m }

// ListRequestValidationError is the validation error returned by
// ListRequest.Validate if the designated constraints aren't met.
type ListRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListRequestValidationError) ErrorName() string { return "ListRequestValidationError" }

// Error satisfies the builtin error interface
func (e ListRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListRequestValidationError{}

// Validate checks the field values on ListResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ListResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ListResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ListResponseMultiError, or
// nil if none found.
func (m *ListResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ListResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return ListResponseMultiError(errors)
	}

	return nil
}

// ListResponseMultiError is an error wrapping multiple validation errors
// returned by ListResponse.ValidateAll() if the designated constraints aren't met.
type ListResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ListResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ListResponseMultiError) AllErrors() []error { return m }

// ListResponseValidationError is the validation error returned by
// ListResponse.Validate if the designated constraints aren't met.
type ListResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListResponseValidationError) ErrorName() string { return "ListResponseValidationError" }

// Error satisfies the builtin error interface
func (e ListResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListResponseValidationError{}
