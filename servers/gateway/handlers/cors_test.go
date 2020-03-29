package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	cases := []struct {
		name               string
		hint               string
		req                *http.Request
		expectedStatusCode int
	}{
		{
			"Successful GET request",
			"Make sure CORS middleware is working",
			httptest.NewRequest("GET", "http://localhost:3000/v1/summary", nil),
			200,
		},
		{
			"Failed GET request",
			"If route does not exist, there should be an error",
			httptest.NewRequest("GET", "http://localhost:3000/DOES_NOT_EXIST/", nil),
			404,
		},
		{
			"Successful POST request",
			"Make sure CORS middleware is working",
			httptest.NewRequest("POST", "http://localhost:3000/v1/summary", nil),
			200,
		},
		{
			"Failed POST request",
			"If route does not exist, there should be an error",
			httptest.NewRequest("POST", "http://localhost:3000/DOES_NOT_EXIST/", nil),
			404,
		},
		{
			"Successful PUT request",
			"Make sure CORS middleware is working",
			httptest.NewRequest("PUT", "http://localhost:3000/v1/summary", nil),
			200,
		},
		{
			"Failed PUT request",
			"If route does not exist, there should be an error",
			httptest.NewRequest("PUT", "http://localhost:3000/DOES_NOT_EXIST/", nil),
			404,
		},
		{
			"Successful PATCH request",
			"Make sure CORS middleware is working",
			httptest.NewRequest("PATCH", "http://localhost:3000/v1/summary", nil),
			200,
		},
		{
			"Failed PATCH request",
			"If route does not exist, there should be an error",
			httptest.NewRequest("PATCH", "http://localhost:3000/DOES_NOT_EXIST/", nil),
			404,
		},
		{
			"Successful DELETE request",
			"Make sure CORS middleware is working",
			httptest.NewRequest("DELETE", "http://localhost:3000/v1/summary", nil),
			200,
		},
		{
			"Failed DELETE request",
			"If route does not exist, there should be an error",
			httptest.NewRequest("DELETE", "http://localhost:3000/DOES_NOT_EXIST/", nil),
			404,
		},
		{
			"Successful OPTIONS request",
			"Make sure CORS middleware is working",
			httptest.NewRequest("OPTIONS", "http://localhost:3000/v1/summary", nil),
			200,
		},
	}

	for _, c := range cases {
		mux := http.NewServeMux()
		rec := httptest.NewRecorder()
		handler := func(w http.ResponseWriter, r *http.Request) {}
		mux.HandleFunc("/v1/summary", handler)
		corsMiddleware := NewCORS(mux)
		corsMiddleware.ServeHTTP(rec, c.req)
		server := httptest.NewServer(corsMiddleware)
		defer server.Close()
		if status := rec.Code; status != c.expectedStatusCode {
			t.Errorf("Error: %s\nHandler returned wrong status code: got %v want %v\nHint:%s",
				c.name, status, c.expectedStatusCode, c.hint)
		}
	}
}
