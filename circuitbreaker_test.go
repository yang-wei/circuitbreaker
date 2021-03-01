package circuitbreaker_test

import (
	"errors"
	"testing"

	"github.com/yang-wei/circuitbreaker"
)

func TestExecuteWhenClose(t *testing.T) {
	cb := circuitbreaker.DefaultCircuitBreaker()
	result, err := cb.Run(func() (interface{}, error) {
		r := 1
		return r, nil
	})

	assertErrorIsNotNil(t, err)
	assertEqual(t, 1, result)
}

func TestTripToOpenWhenExceedFailureThreshold(t *testing.T) {
	cb := circuitbreaker.DefaultCircuitBreaker()

	myErr := errors.New("my error")
	_, firstTryErr := cb.Run(func() (interface{}, error) {
		return nil, myErr
	})
	assertEqual(t, firstTryErr, myErr)

	_, secondTryErr := cb.Run(func() (interface{}, error) {
		return nil, nil
	})
	assertEqual(t, secondTryErr, circuitbreaker.ErrOpened)
}

func assertErrorIsNotNil(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expect err to be nil but got %v", err)
	}
}

func assertEqual(t *testing.T, want, got interface{}) {
	if want != got {
		t.Errorf("Want %v but got %v", want, got)
	}
}
