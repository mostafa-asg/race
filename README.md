# Race
Package race provides functionality to call http requests simultaneously and get the result from the fastest one.

## Usage
```Go
req1, err := http.NewRequest("GET", "slowServer.url", nil)
req2, err := http.NewRequest("GET", "fastServer.url", nil)
req3, err := http.NewRequest("GET", "backupServer.url", nil)

// Starts all 3 requests concurrently and returns
// the *http.Response got from fastest server
res, err := race.Between(req1, req2, req3)
```
Or if you prefer starting a request first and only if it takes too long, starts the other, use `FirstThenStart`:
```Go
// First start `req1` and after 1 second start the other requests (req2 and req3)
res, err := race.FirstThenStart(req1, 1*time.Second, req2, req3)
```
