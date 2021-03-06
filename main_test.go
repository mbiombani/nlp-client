package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
)

func TestGetEnvSet(t *testing.T) {
	err := os.Setenv("API_KEY", "foo")
	if err != nil {
		t.Error(err)
	}
	if assert.Equal(t, getEnv("API_KEY", "bar"), "foo") {
	}
}

func TestGetEnvNotSet(t *testing.T) {
	err := os.Unsetenv("API_KEY")
	if err != nil {
		t.Error(err)
	}
	if assert.Equal(t, getEnv("API_KEY", "bar"), "bar") {
	}
}

func TestGetError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)
	res := w.Result()
	err := res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	expected := `code=500, message=Internal Server Error`
	if assert.EqualError(t, getError(c), expected) {
	}
}

func TestGetHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)
	err := getHealth(c)
	if err != nil {
		t.Error(err)
	}
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Error(err)
		}
	}(res.Body)
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	type HealthStatus struct {
		Status string `json:"status"`
	}
	expectedStatus := HealthStatus{"Up"}
	var responseBody HealthStatus

	err = json.Unmarshal(data, &responseBody)
	if err != nil {
		t.Error(err)
	}

	if assert.NoError(t, getHealth(c)) {
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, &expectedStatus, &responseBody)
	}
}

func TestGetRoutes(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/routes", nil)
	w := httptest.NewRecorder()
	e.GET("/health", getHealth)
	e.GET("/health/:app", getHealthUpstream)
	e.GET("/error", getError)
	e.GET("/routes", getRoutes)
	e.POST("/keywords", getKeywords)
	e.POST("/tokens", getTokens)
	e.POST("/entities", getEntities)
	e.POST("/sentences", getSentences)
	e.POST("/language", getLanguage)
	e.POST("/record", putDynamo)
	c := e.NewContext(req, w)
	err := getRoutes(c)
	if err != nil {
		t.Error(err)
	}
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Error(err)
		}
	}(res.Body)
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	type Route struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Name   string `json:"name"`
	}

	prefix := "github.com/garystafford/nlp-client"
	expectedStatus := []Route{
		{"GET", "/health", prefix + ".getHealth"},
		{"GET", "/health/:app", prefix + ".getHealthUpstream"},
		{"GET", "/error", prefix + ".getError"},
		{"GET", "/routes", prefix + ".getRoutes"},
		{"POST", "/keywords", prefix + ".getKeywords"},
		{"POST", "/tokens", prefix + ".getTokens"},
		{"POST", "/entities", prefix + ".getEntities"},
		{"POST", "/sentences", prefix + ".getSentences"},
		{"POST", "/language", prefix + ".getLanguage"},
		{"POST", "/record", prefix + ".putDynamo"},
	}
	var responseBody []Route

	err = json.Unmarshal(data, &responseBody)
	if err != nil {
		t.Error(err)
	}

	// sort both arrays of Route structs so they are in identical order
	sort.Slice(expectedStatus, func(i, j int) bool { return expectedStatus[i].Path < expectedStatus[j].Path })
	sort.Slice(responseBody, func(i, j int) bool { return responseBody[i].Path < responseBody[j].Path })

	if assert.NoError(t, getRoutes(c)) {
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, &expectedStatus, &responseBody)
	}
}

func TestGetHealthUpstreamRake(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/rake", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)
	res := w.Result()
	res.Body.Close()

	expected := `code=405, message=Method Not Allowed`
	if assert.EqualError(t, getHealthUpstream(c), expected) {
	}
}

func TestGetKeywords(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/keywords", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)
	res := w.Result()
	res.Body.Close()

	expected := `code=500, message=Post "http://localhost:8081/keywords": dial tcp [::1]:8081: connect: connection refused`
	if assert.EqualError(t, getKeywords(c), expected) {
	}
}
