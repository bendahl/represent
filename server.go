package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func startServer(inputMd string, host string, port int) error {
	inFile, err := os.Open(inputMd)
	if err != nil {
		return err
	}
	defer inFile.Close()
	slides, err := split(io.Reader(inFile))
	if err != nil {
		return err
	}
	libs, err := fs.Sub(templateFs, "libs")
	if err != nil {
		return err
	}
	slideNum := 1
	var m sync.RWMutex
	bc := NewBroadcaster()
	http.Handle("/libs/", http.StripPrefix("/libs/", http.FileServer(http.FS(libs))))
	http.Handle("/assets/", http.FileServer(http.Dir(filepath.Dir(inputMd))))
	http.HandleFunc("/slides", slidesHandler(slides, false, &slideNum, &m))
	http.HandleFunc("/presenter", slidesHandler(slides, true, &slideNum, &m))
	http.HandleFunc("/next", nextHandler(slides, &slideNum, len(slides), &m, bc))
	http.HandleFunc("/prev", prevHandler(slides, &slideNum, len(slides), &m, bc))
	http.HandleFunc("/events", sseHanlder(bc))
	listenAddr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("server listening on %s\n", listenAddr)
	fmt.Printf("presenter view available at: http://%s/presenter\n", listenAddr)
	fmt.Printf("slides view available at: http://%s/slides\n", listenAddr)

	return http.ListenAndServe(listenAddr, nil)
}

func sseHanlder(events *Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Accel-Buffering", "no")
		rc := http.NewResponseController(w)
		ch := events.Subscribe()
		defer events.Unsubscribe(ch)

		for {
			select {
			case <-r.Context().Done():
				return

			case slideNum := <-ch:
				_, err := fmt.Fprintf(
					w,
					"event: slideChanged\ndata: %d\n\n",
					slideNum,
				)
				if err != nil {
					return
				}

				if err := rc.Flush(); err != nil {
					return
				}
			}
		}
	}
}

func changeSlide(slides []Slide, slideNum *int, totalSlides int, m *sync.RWMutex, bc *Broadcaster, delta int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		m.Lock()
		*slideNum += delta
		*slideNum = min(max(1, *slideNum), totalSlides)
		current := *slideNum
		m.Unlock()
		bc.Publish(current)
		updateSlide(w, slides, current, true)
	}
}

func nextHandler(slides []Slide, slideNum *int, totalSlides int, m *sync.RWMutex, bc *Broadcaster) http.HandlerFunc {
	return changeSlide(slides, slideNum, totalSlides, m, bc, 1)
}

func prevHandler(slides []Slide, slideNum *int, totalSlides int, m *sync.RWMutex, bc *Broadcaster) http.HandlerFunc {
	return changeSlide(slides, slideNum, totalSlides, m, bc, -1)
}

func slidesHandler(slides []Slide, isPresenter bool, slideNum *int, m *sync.RWMutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.RLock()
		current := *slideNum
		m.RUnlock()
		updateSlide(w, slides, current, isPresenter)
	}
}
func updateSlide(w http.ResponseWriter, slides []Slide, slideNum int, isPresenter bool) {
	slide, err := renderSlide(slides[slideNum-1], isPresenter, slideNum, len(slides))
	if err != nil {
		w.Write([]byte(slide))
		w.WriteHeader(http.StatusInternalServerError)
		// Write body after setting header
		_, _ = fmt.Fprint(w, "Failed to render slide")
		return
	}
	_, _ = w.Write([]byte(slide))
}
