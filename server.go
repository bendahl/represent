package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type presentationServer struct {
	slides []Slide
}

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
	http.Handle("/libs/", http.StripPrefix("/libs/", http.FileServer(http.FS(libs))))
	http.Handle("/assets/", http.FileServer(http.Dir(filepath.Dir(inputMd))))
	http.HandleFunc("/slides", slidesHandlerProducer(slides))

	listenAddr := "localhost:8080"
	fmt.Printf("server listening on %s\n", listenAddr)

	return http.ListenAndServe(listenAddr, nil)
}

func slidesHandlerProducer(slides []Slide) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		qParams := r.URL.Query()
		var err error
		isPresenter := false
		if qParams.Has("presenter") {
			isPresenter, err = strconv.ParseBool(qParams.Get("presenter"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprint(w, "Invalid presenter query parameter (not a boolean)")
				return
			}
		}
		slideNum := 1
		if qParams.Has("slide") {
			slideNum, err = strconv.Atoi(qParams["slide"][0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprint(w, "Invalid slide number query parameter")
				return
			}
		}
		slideNum = max(1, slideNum)
		if r.Header.Get("HX-Request") == "true" {
			target := fmt.Sprintf(
				"/slides?presenter=%t&slide=%d",
				isPresenter,
				slideNum,
			)
			w.Header().Set("HX-Redirect", target)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		slide, err := renderSlide(slides[slideNum-1], isPresenter, slideNum, len(slides))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			// Write body after setting header
			_, _ = fmt.Fprint(w, "Failed to render slide")
			return
		}
		_, _ = w.Write([]byte(slide))
	}
}
