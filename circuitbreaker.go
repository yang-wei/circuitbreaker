package circuitbreaker

import (
	"errors"
	"time"
)

type State int

const (
	close State = iota
	open
	halfOpen
)

var (
	ErrOpen = errors.New("Circuit breaker is opened")
)

type Counter struct {
	success int
	failure int
	started int64
}

func (c *Counter) reset() {
	c.success = 0
	c.failure = 0
	c.started = now()
}

func (c *Counter) recordSuccess() {
	c.success = c.success + 1
}

func (c *Counter) recordFailure() {
	c.failure = c.failure + 1
}

func (c *Counter) successRate() float64 {
	return float64(c.success) / float64(c.success+c.failure)
}

func (c *Counter) failureRate() float64 {
	return float64(c.failure) / float64(c.success+c.failure)
}

func now() int64 {
	return time.Now().UnixNano()
}

func NewCounter() *Counter {
	return &Counter{
		success: 0,
		failure: 0,
		started: now(),
	}
}

type CircuitBreaker struct {
	state                State
	counter              *Counter
	samplingPeriodInMs   int64   // counter reset when sampling period is over
	failureRateThreshold float64 // trip breaker from CLOSED to OPENED when threshold exceed
	successRateThreshold float64 // trip breaker from OPENED to CLOSED when threshold exceed
	halfOpenTimeoutMs    int64   // trip breaker from OPENED to HALF-OPENED after timeout exceed
}

func New(samplingPeriodInMs int64, failureRateThreshold, successRateThreshold float64, halfOpenTimeoutMs int64) *CircuitBreaker {
	return &CircuitBreaker{
		state:              close,
		counter:            NewCounter(),
		samplingPeriodInMs: samplingPeriodInMs,
	}
}

func (cb *CircuitBreaker) Run(runnable func() (interface{}, error)) (interface{}, error) {
	cb.evaluteState()
	if cb.state == open {
		return nil, ErrOpen
	}
	result, err := runnable()
	cb.recordStat(err)
	return result, err
}

func (cb *CircuitBreaker) recordStat(err error) {
	if now()-cb.counter.started > (cb.samplingPeriodInMs * 1000) {
		cb.counter.reset()
	}
	if err == nil {
		cb.counter.recordSuccess()
	} else {
		cb.counter.recordFailure()
	}
	cb.evaluteState()
}

func (cb *CircuitBreaker) evaluteState() {
	if cb.state == open && now()-cb.counter.started > (cb.halfOpenTimeoutMs*1000) {
		cb.changeState(halfOpen)
	} else if cb.state == close {
		// might go to OPEN
		if cb.counter.failureRate() > cb.failureRateThreshold {
			cb.changeState(open)
		}
	} else {
		// half open: might go to either CLOSE or OPEN
		// TODO: state might stay half open
		if cb.counter.successRate() > cb.successRateThreshold {
			cb.changeState(close)
		}
		if cb.counter.failureRate() > cb.failureRateThreshold {
			cb.changeState(open)
		}
	}
}

func (cb *CircuitBreaker) changeState(s State) {
	cb.state = s
	cb.counter.reset()
}
