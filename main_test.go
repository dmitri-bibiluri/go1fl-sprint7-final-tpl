package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	city := "moscow"
	total := len(cafeList[city])

	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: min(100, total)},
	}

	for _, tc := range requests {
		t.Run("count="+strconv.Itoa(tc.count), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/cafe?city="+city+"&count="+strconv.Itoa(tc.count), nil)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := strings.TrimSpace(response.Body.String())

			got := 0
			if body != "" {
				items := strings.Split(body, ",")
				got = len(items)
			}

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	city := "moscow"
	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, tc := range requests {
		t.Run("search="+tc.search, func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/cafe?city="+city+"&search="+tc.search, nil)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := strings.TrimSpace(response.Body.String())
			if body == "" {
				assert.Equal(t, tc.wantCount, 0)
				return
			}

			items := strings.Split(body, ",")
			assert.Equal(t, tc.wantCount, len(items))

			needle := strings.ToLower(tc.search)
			for _, item := range items {
				hay := strings.ToLower(strings.TrimSpace(item))
				assert.True(t, strings.Contains(hay, needle), "кафе %q не содержит подстроку %q", item, tc.search)
			}
		})
	}
}