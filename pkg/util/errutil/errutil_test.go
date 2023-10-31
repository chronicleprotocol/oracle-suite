package errutil

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppend(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errors.New("error2")
	multiErr := MultiError{err1, err2}

	t.Run("no errors", func(t *testing.T) {
		result := Append(nil)
		assert.Nil(t, result)
	})

	t.Run("single error", func(t *testing.T) {
		result := Append(err1)
		assert.Equal(t, err1, result)
	})

	t.Run("multiple errors", func(t *testing.T) {
		result := Append(err1, err2)
		assert.IsType(t, MultiError{}, result)
		assert.Contains(t, result.(MultiError), err1)
		assert.Contains(t, result.(MultiError), err2)
	})

	t.Run("append MultiError to error", func(t *testing.T) {
		result := Append(err1, multiErr)
		assert.IsType(t, MultiError{}, result)
		assert.Contains(t, result.(MultiError), err1)
		assert.Contains(t, result.(MultiError), err2)
	})

	t.Run("append error to MultiError", func(t *testing.T) {
		result := Append(multiErr, err1)
		assert.IsType(t, MultiError{}, result)
		assert.Contains(t, result.(MultiError), err1)
		assert.Contains(t, result.(MultiError), err2)
	})

	t.Run("append MultiError to MultiError", func(t *testing.T) {
		result := Append(multiErr, multiErr)
		assert.IsType(t, MultiError{}, result)
		assert.Contains(t, result.(MultiError), err1)
		assert.Contains(t, result.(MultiError), err2)
		assert.Len(t, result.(MultiError), 4) // It should have 4 errors since we appended the same multiError.
	})
}

func TestMultiError(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errors.New("error2")

	t.Run("Empty MultiError", func(t *testing.T) {
		var multiErr MultiError
		assert.Empty(t, multiErr.Error())
	})

	t.Run("Single error MultiError", func(t *testing.T) {
		multiErr := MultiError{err1}
		assert.Equal(t, "following errors occurred: [error1]", multiErr.Error())
	})

	t.Run("Multiple errors MultiError", func(t *testing.T) {
		multiErr := MultiError{err1, err2}
		assert.Equal(t, "following errors occurred: [error1, error2]", multiErr.Error())
	})
}

func TestIgnore(t *testing.T) {
	tests := []struct {
		fn    func() (int, error)
		value int
	}{
		{
			fn: func() (int, error) {
				return 1, nil
			},
			value: 1,
		},
		{
			fn: func() (int, error) {
				return 1, fmt.Errorf("error")
			},
			value: 1,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			assert.Equal(t, tt.value, Ignore(tt.fn()))
		})
	}
}

func TestMust(t *testing.T) {
	tests := []struct {
		fn    func() (int, error)
		panic bool
		value int
	}{
		{
			fn: func() (int, error) {
				return 1, nil
			},
			panic: false,
			value: 1,
		},
		{
			fn: func() (int, error) {
				return 1, fmt.Errorf("error")
			},
			panic: true,
			value: 1,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			defer func() {
				assert.Equal(t, tt.panic, recover() != nil)
			}()
			assert.Equal(t, tt.value, Must(tt.fn()))
		})
	}
}
