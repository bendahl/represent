package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func startServer(inputMd string) error {
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
	var m sync.Mutex
	http.Handle("/libs/", http.StripPrefix("/libs/", http.FileServer(http.FS(libs))))
	http.Handle("/assets/", http.FileServer(http.Dir(filepath.Dir(inputMd))))
	http.HandleFunc("/slides", slidesHandlerProducer(slides, false, &slideNum, &m))
	http.HandleFunc("/presenter", slidesHandlerProducer(slides, true, &slideNum, &m))
	http.HandleFunc("/next", slidesHandlerProducer(slides, true, &slideNum, &m))
	http.HandleFunc("/prev", slidesHandlerProducer(slides, true, &slideNum, &m))

	listenAddr := "localhost:8080"
	fmt.Printf("server listening on %s\n", listenAddr)

	return http.ListenAndServe(listenAddr, nil)
}

func slidesHandlerProducer(slides []Slide, isPresenter bool, slideNum *int, m *sync.Mutex) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		m.Lock()
		defer m.Unlock()
		switch true {
		case strings.HasPrefix(r.URL.Path, "/next"):
			*slideNum += 1
		case strings.HasPrefix(r.URL.Path, "/prev"):
			*slideNum -= 1
		}
		*slideNum = min(max(1, *slideNum), len(slides))
		slide, err := renderSlide(slides[*slideNum-1], isPresenter, *slideNum, len(slides))
		if err != nil {
			w.Write([]byte(slide))
			w.WriteHeader(http.StatusInternalServerError)
			// Write body after setting header
			_, _ = fmt.Fprint(w, "Failed to render slide")
			return
		}
		_, _ = w.Write([]byte(slide))
	}
}
