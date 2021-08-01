package race

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-multierror"
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
	onError := make(chan error)

	// run all the requests concurrently
	for _, r := range reqs {
		req := r.WithContext(ctx)
		go race.makeRequest(onComplete, onError, req)
	}

	var errs []error
	for {
		select {
		case res := <-onComplete:
			return res, nil
		case err := <-onError:
			errs = append(errs, err)

			// all requests failed
			if len(errs) == len(reqs) {
				allerrors := &multierror.Error{}
				multierror.Append(allerrors, errs...)
				return nil, allerrors
			}
		}
	}
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
// if all requests failed, it will return *multierror.Error containing all errors that happened
func Between(reqs ...*http.Request) (*http.Response, error) {
	return New().Between(reqs...)
}

// BetweenWithClient is like Between but gets user's http client
func BetweenWithClient(client *http.Client, reqs ...*http.Request) (*http.Response, error) {
	return NewWithClient(client).Between(reqs...)
}

func (race *Race) makeRequest(onComplete chan *http.Response, onError chan error, req *http.Request) {
	res, err := race.client.Do(req)
	if err != nil {
		onError <- err
		return
	}

	onComplete <- res
}

func createContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(context.Background(), timeout)
	}

	return context.WithCancel(context.Background())
}
