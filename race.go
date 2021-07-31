package race

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Race between requests
type Race struct {
	client *http.Client
}

// Between gets a bunch of requests and makes http request simultaneously to all of them
// the first answer will be returned
func (race *Race) Between(reqs ...*http.Request) (*http.Response, error) {
	ctx, cancel := createContext(race.client.Timeout)
	defer cancel()

	onComplete := make(chan *http.Response)

	// run all the requests concurrently
	for _, r := range reqs {
		req := r.WithContext(ctx)
		go race.makeRequest(onComplete, req)
	}

	if race.client.Timeout > 0 {
		for {
			select {
			case res := <-onComplete:
				return res, nil
			case <-ctx.Done():
				return nil, errors.New("Timeout")
			}
		}
	}

	return <-onComplete, nil
}

// New returns new race object with default http client
func New() *Race {
	return NewWithClient(http.DefaultClient)
}

// NewWithClient returns new race object with the given http client
func NewWithClient(client *http.Client) *Race {
	return &Race{
		client: client,
	}
}

// Between gets a bunch of requests and makes http request simultaneously to all of them
// the first answer will be returned
func Between(reqs ...*http.Request) (*http.Response, error) {
	return New().Between(reqs...)
}

func (race *Race) makeRequest(onComplete chan *http.Response, req *http.Request) {
	res, err := race.client.Do(req)
	if err == nil {
		onComplete <- res
	}
}

func createContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(context.Background(), timeout)
	}

	return context.WithCancel(context.Background())
}
