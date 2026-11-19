package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() *Server {
	slides := []Slide{
		{
			IsFrontPage: true,
			Title:       "Test Presentation",
			Content:     []string{"Test subtitle"},
			Notes:       []string{"Introduction notes"},
		},
		{
			Title:   "Second Slide",
			Content: []string{"Second slide content"},
			Notes:   []string{"Second slide notes"},
		},
	}

	return NewServer(slides, "light", "light")
}

func TestServerViewer(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/slides", nil)
	rec := httptest.NewRecorder()

	srv.viewer(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want HTML content type", got)
	}

	body := rec.Body.String()

	if !strings.Contains(body, "Test Presentation") {
		t.Error("viewer response should contain presentation title")
	}

	if !strings.Contains(body, "Test subtitle") {
		t.Error("viewer response should contain subtitle")
	}

	if !strings.Contains(body, `id="footer" hidden`) {
		t.Error("viewer response should hide presenter notes")
	}
}

func TestServerPresenter(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/presenter", nil)
	rec := httptest.NewRecorder()

	srv.presenter(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()

	if !strings.Contains(body, "Test Presentation") {
		t.Error("presenter response should contain presentation title")
	}

	if !strings.Contains(body, "Introduction notes") {
		t.Error("presenter response should contain presenter notes")
	}
	if !strings.Contains(body, `id="footer"`) {
		t.Error("presenter response should contain notes footer")
	}

	if strings.Contains(body, `id="footer" hidden`) {
		t.Error("presenter notes should not be hidden")
	}

	if !strings.Contains(body, "Introduction notes") {
		t.Error("presenter response should contain presenter notes")
	}
}

func TestServerNavigation(t *testing.T) {
	tests := []struct {
		name     string
		start    int
		delta    int
		expected int
	}{
		{"next slide", 1, 1, 2},
		{"previous slide", 2, -1, 1},
		{"cannot advance beyond last slide", 2, 1, 2},
		{"cannot go before first slide", 1, -1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer()
			srv.slideNum = tt.start

			got := srv.changeSlide(tt.delta)

			if got != tt.expected {
				t.Errorf("changeSlide(%d) = %d, want %d",
					tt.delta, got, tt.expected)
			}

			if got := srv.currentSlide(); got != tt.expected {
				t.Errorf("currentSlide() = %d, want %d",
					got, tt.expected)
			}
		})
	}
}

func TestServerNextAndPrevious(t *testing.T) {
	srv := newTestServer()

	t.Run("next", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/next", nil)
		rec := httptest.NewRecorder()

		srv.next(rec, req)

		if got := srv.currentSlide(); got != 2 {
			t.Errorf("currentSlide() = %d, want 2", got)
		}

		if !strings.Contains(rec.Body.String(), "Second Slide") {
			t.Error("next response should contain second slide")
		}
	})

	t.Run("previous", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/prev", nil)
		rec := httptest.NewRecorder()

		srv.prev(rec, req)

		if got := srv.currentSlide(); got != 1 {
			t.Errorf("currentSlide() = %d, want 1", got)
		}

		if !strings.Contains(rec.Body.String(), "Test Presentation") {
			t.Error("previous response should contain first slide")
		}
	})
}

func TestServerEventsHandler(t *testing.T) {
	srv := newTestServer()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	srv.eventsHandler(rec, req)

	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", got)
	}

	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", got)
	}

	if got := rec.Header().Get("X-Accel-Buffering"); got != "no" {
		t.Errorf("X-Accel-Buffering = %q, want no", got)
	}
}
