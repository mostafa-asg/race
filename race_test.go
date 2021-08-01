package race

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
)

func TestBetweenSlowAndFast(t *testing.T) {
	slow := []byte("slow")
	fast := []byte("fast")

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.Write(slow)
	}))
	defer slowServer.Close()

	fastServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Microsecond)
		w.Write(fast)
	}))
	defer fastServer.Close()

	req1, err := http.NewRequest("GET", slowServer.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	req2, err := http.NewRequest("GET", fastServer.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Between(req1, req2)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	bytes.Compare(resBytes, fast)
}

func TestBetweenUnresolvableAndResolvableHost(t *testing.T) {
	hello := []byte("hello")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.Write(hello)
	}))
	defer server.Close()

	req1, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	req2, err := http.NewRequest("GET", "http://CrazyAndStrangeAndUnresolvableDomain", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Between(req2, req1)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	bytes.Compare(resBytes, hello)
}

func TestBetweenFailAndTimeoutReq(t *testing.T) {
	hello := []byte("hello")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write(hello)
	}))
	defer server.Close()

	req1, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	req2, err := http.NewRequest("GET", "http://CrazyAndStrangeAndUnresolvableDomain", nil)
	if err != nil {
		t.Fatal(err)
	}

	r := NewWithClient(&http.Client{
		Timeout: 1 * time.Second,
	})
	res, err := r.Between(req1, req2)
	if res != nil {
		t.Fatal(err)
	}

	multiError, ok := err.(*multierror.Error)
	if !ok {
		t.Fatal("Expected error of type *multierror.Error")
	}

	if len(multiError.Errors) != 2 {
		t.Fatal("Expected 2 errors")
	}
}

func TestBetweenAllFailedReq(t *testing.T) {
	req1, err := http.NewRequest("GET", "http://CrazyAndStrangeAndUnresolvableDomain", nil)
	if err != nil {
		t.Fatal(err)
	}

	req2, err := http.NewRequest("GET", "http://CrazyAndStrangeAndUnresolvableDomain", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Between(req1, req2)
	if res != nil {
		t.Fatal(err)
	}

	multiError, ok := err.(*multierror.Error)
	if !ok {
		t.Fatal("Expected error of type *multierror.Error")
	}

	if len(multiError.Errors) != 2 {
		t.Fatal("Expected 2 errors")
	}
}
