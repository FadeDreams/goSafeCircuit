## goSafeCircuit

`gosafecircuit` is a Go package that provides a circuit breaker implementation to manage and control the execution of functions based on their failure rates.

### Overview

The `CircuitBreaker` struct represents a circuit breaker that monitors the number of consecutive failures of a function and controls whether to allow execution based on predefined thresholds.

### Installation

To use `goSafeCircuit`, you need to have Go installed. Use the following command to install the package:

```bash
go get -u github.com/your-username/goSafeCircuit
```

### Example Usage

```go
package main

import (
	"fmt"
	goSafeCircuit "github.com/fadedreams/gosafecircuit"
	"time"
)

func main() {
	// Create a new circuit breaker with maxFailures=3 and timeout=5 seconds and pauseTime=1 second
	cb := goSafeCircuit.NewCircuitBreaker(3, 5*time.Second, 1*time.Second)

	// Example function that might fail
	exampleFunction := func() error {
		// Simulate some operation that may fail
		if time.Now().Unix()%2 == 0 {
			return fmt.Errorf("an error occurred")
		}
		return nil
	}

	// Execute the function with circuit breaker protection
	for i := 0; i < 10; i++ {
		err := cb.Execute(exampleFunction)
		if err != nil {
			fmt.Printf("Attempt %d failed: %v\n", i+1, err)
		} else {
			fmt.Printf("Attempt %d succeeded\n", i+1)
		}
		time.Sleep(1 * time.Second) // Simulate some delay between attempts
	}
}

```
