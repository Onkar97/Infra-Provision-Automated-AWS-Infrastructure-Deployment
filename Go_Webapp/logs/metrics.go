package logs

import (
	"log"
	"os"

	"github.com/alexcesaro/statsd"
)

// ClientInterface defines the methods our app uses.
// This interface allows us to swap the Real client for a Mock client easily.
type ClientInterface interface {
	Increment(bucket string)
	Timing(bucket string, value interface{})
	Close()
}

// Client is the global instance other packages will use.
// e.g., metrics.Client.Increment("my.counter")
var Client ClientInterface

// NoOpClient is our "Test" client that does nothing.
type NoOpClient struct{}

func (n *NoOpClient) Increment(bucket string)                 {}
func (n *NoOpClient) Timing(bucket string, value interface{}) {}
func (n *NoOpClient) Close()                                  {}

// Init initializes the metrics client based on the environment.
// This should be called once in your main.go.
func Init() {
	env := os.Getenv("APP_ENV") // or GO_ENV

	// 1. If in Test mode, use the NoOp (Dummy) client
	if env == "test" {
		Client = &NoOpClient{}
		return
	}

	// 2. Setup Real Client configuration
	host := os.Getenv("STATSD_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	address := host + ":8125"

	// 3. Create the client with an Error Handler
	// This mirrors your `socket.on('error')` logic.
	c, err := statsd.New(
		statsd.Address(address),
		statsd.ErrorHandler(func(err error) {
			// Use standard log to avoid circular dependency with your custom logger
			log.Printf("CRITICAL: Error connecting to StatsD: %v", err)
		}),
	)

	// If the client fails to initialize immediately, fallback to NoOp
	// to prevent the entire application from crashing.
	if err != nil {
		log.Printf("Failed to create StatsD client: %v. Falling back to NoOp.", err)
		Client = &NoOpClient{}
		return
	}

	Client = c
}
