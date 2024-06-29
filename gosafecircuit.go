package gosafecircuit

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreaker represents a circuit breaker.
type CircuitBreaker struct {
	mutex               sync.Mutex
	state               State
	consecutiveFailures int
	maxFailures         int
	timeout             time.Duration
	openTimeout         time.Time
	pauseTime           time.Duration // Added pause time between retries in HALF-OPEN state
}

// State represents the state of the circuit breaker.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker creates a new CircuitBreaker instance.
func NewCircuitBreaker(maxFailures int, timeout time.Duration, pauseTime time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
		openTimeout: time.Time{},
		pauseTime:   pauseTime,
	}
}

// Execute executes the given function with circuit breaker logic.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case StateOpen:
		if time.Now().After(cb.openTimeout) {
			cb.state = StateHalfOpen
		} else {
			return errors.New("circuit breaker is open")
		}
	case StateHalfOpen:
		// Try the operation
		err := fn()
		if err == nil {
			cb.reset()
		} else {
			cb.trip()
		}
		time.Sleep(cb.pauseTime) // Pause before next try in HALF-OPEN state
		return err
	}

	// Execute the operation
	err := fn()
	if err == nil {
		cb.reset()
	} else {
		cb.trip()
	}
	return err
}

// trip trips the circuit breaker to the open state.
func (cb *CircuitBreaker) trip() {
	cb.state = StateOpen
	cb.consecutiveFailures = 0
	cb.openTimeout = time.Now().Add(cb.timeout)
}

// reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) reset() {
	cb.state = StateClosed
	cb.consecutiveFailures = 0
}
