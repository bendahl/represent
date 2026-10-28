package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type Server struct {
	slides      []Slide
	theme       string
	hljsTheme   string
	mu          sync.RWMutex
	slideNum    int
	broadcaster *Broadcaster
}

func NewServer(slides []Slide, theme string, hljsTheme string) *Server {
	return &Server{
		slides:      slides,
		theme:       theme,
		hljsTheme:   hljsTheme,
		slideNum:    1,
		broadcaster: NewBroadcaster(),
	}
}

func startServer(inputMd string, host string, port int, theme string, hljsTheme string) error {
	inFile, err := os.Open(inputMd)
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer inFile.Close()

	slides, err := split(io.Reader(inFile))
	if err != nil {
		return fmt.Errorf("split slides: %w", err)
	}
	if len(slides) == 0 {
		return fmt.Errorf("input contains no slides")
	}

	libs, err := fs.Sub(templateFs, "libs")
	if err != nil {
		return fmt.Errorf("load libraries filesystem: %w", err)
	}

	srv := NewServer(slides, theme, hljsTheme)
	mux := http.NewServeMux()
	mux.Handle("/libs/", http.StripPrefix("/libs/", http.FileServer(http.FS(libs))))
	mux.Handle("/assets/", http.FileServer(http.Dir(filepath.Dir(inputMd))))
	mux.HandleFunc("/slides", srv.viewer)
	mux.HandleFunc("/presenter", srv.presenter)
	mux.HandleFunc("/next", srv.next)
	mux.HandleFunc("/prev", srv.prev)
	mux.HandleFunc("/events", srv.eventsHandler)

	listenAddr := fmt.Sprintf("%s:%d", host, port)

	log.Printf("server listening on %s", listenAddr)
	log.Printf("presenter view available at: http://%s/presenter", listenAddr)
	log.Printf("slides view available at: http://%s/slides", listenAddr)

	return http.ListenAndServe(listenAddr, mux)
}

func (s *Server) currentSlide() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.slideNum
}

func (s *Server) changeSlide(delta int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slideNum = min(max(1, s.slideNum+delta), len(s.slides))
	return s.slideNum
}

func (s *Server) viewer(w http.ResponseWriter, r *http.Request) {
	s.serveSlide(w, s.currentSlide(), false)
}

func (s *Server) presenter(w http.ResponseWriter, r *http.Request) {
	s.serveSlide(w, s.currentSlide(), true)
}

func (s *Server) next(w http.ResponseWriter, r *http.Request) {
	s.changeSlideHandler(w, 1)
}

func (s *Server) prev(w http.ResponseWriter, r *http.Request) {
	s.changeSlideHandler(w, -1)
}

func (s *Server) changeSlideHandler(w http.ResponseWriter, delta int) {
	current := s.changeSlide(delta)

	s.broadcaster.Publish(current)
	s.serveSlide(w, current, true)
}

func (s *Server) serveSlide(w http.ResponseWriter, slideNum int, isPresenter bool) {
	slide, err := renderSlide(s.slides[slideNum-1], isPresenter, slideNum, len(s.slides), s.theme, s.hljsTheme)
	if err != nil {
		log.Printf("render slide %d: %v", slideNum, err)
		http.Error(w, "Failed to render slide", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, slide); err != nil {
		log.Printf("write slide %d response: %v", slideNum, err)
	}
}

func (s *Server) eventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	rc := http.NewResponseController(w)
	ch := s.broadcaster.Subscribe()
	defer s.broadcaster.Unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return

		case slideNum, ok := <-ch:
			if !ok {
				return
			}

			if _, err := fmt.Fprintf(w, "event: slideChanged\ndata: %d\n\n", slideNum); err != nil {
				return
			}

			if err := rc.Flush(); err != nil {
				return
			}
		}
	}
}
