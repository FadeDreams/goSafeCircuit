package gosafecircuit

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreaker represents a circuit breaker.
type CircuitBreaker struct {
	mutex                   sync.Mutex
	state                   State
	consecutiveFailures     int
	totalFailures           int
	totalSuccesses          int
	maxFailures             int
	timeout                 time.Duration
	openTimeout             time.Time
	pauseTime               time.Duration // Added pause time between retries in HALF-OPEN state
	consecutiveSuccesses    int
	maxConsecutiveSuccesses int
	onOpen                  func()
	onClose                 func()
	onHalfOpen              func()
}

// State represents the state of the circuit breaker.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker creates a new CircuitBreaker instance.
func NewCircuitBreaker(maxFailures int, timeout time.Duration, pauseTime time.Duration, maxConsecutiveSuccesses int) *CircuitBreaker {
	return &CircuitBreaker{
		state:                   StateClosed,
		maxFailures:             maxFailures,
		timeout:                 timeout,
		openTimeout:             time.Time{},
		pauseTime:               pauseTime,
		maxConsecutiveSuccesses: maxConsecutiveSuccesses,
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
			if cb.onHalfOpen != nil {
				cb.onHalfOpen()
			}
		} else {
			return errors.New("circuit breaker is open")
		}
	case StateHalfOpen:
		// Try the operation
		err := fn()
		if err == nil {
			cb.consecutiveSuccesses++
			cb.totalSuccesses++
			if cb.consecutiveSuccesses >= cb.maxConsecutiveSuccesses {
				cb.reset()
			}
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
		cb.totalSuccesses++
	} else {
		cb.consecutiveFailures++
		cb.totalFailures++
		if cb.consecutiveFailures >= cb.maxFailures {
			cb.trip()
		}
	}
	return err
}

// trip trips the circuit breaker to the open state.
func (cb *CircuitBreaker) trip() {
	cb.state = StateOpen
	cb.consecutiveFailures = 0
	cb.consecutiveSuccesses = 0
	cb.openTimeout = time.Now().Add(cb.timeout)
	if cb.onOpen != nil {
		cb.onOpen()
	}
}

// reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) reset() {
	cb.state = StateClosed
	cb.consecutiveFailures = 0
	cb.consecutiveSuccesses = 0
	if cb.onClose != nil {
		cb.onClose()
	}
}

// SetOnOpen sets the callback for when the circuit breaker opens.
func (cb *CircuitBreaker) SetOnOpen(callback func()) {
	cb.onOpen = callback
}

// SetOnClose sets the callback for when the circuit breaker closes.
func (cb *CircuitBreaker) SetOnClose(callback func()) {
	cb.onClose = callback
}

// SetOnHalfOpen sets the callback for when the circuit breaker transitions to half-open.
func (cb *CircuitBreaker) SetOnHalfOpen(callback func()) {
	cb.onHalfOpen = callback
}
