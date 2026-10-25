package main

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

//go:embed libs/*
var templateFs embed.FS

//go:embed slide.html
var slideHtml []byte

type Slide struct {
	IsFrontPage bool
	Title       string
	Content     []string
	Notes       []string
}

func (s Slide) ContentAsHtml() string {
	mergedLines := strings.Join(s.Content, "\n")
	return renderHTML(mergedLines)
}

func (s Slide) NotesAsHtml() string {
	mergedLines := strings.Join(s.Notes, "\n")
	return renderHTML(mergedLines)
}

func generateSlides(outDir string, inputMd string, theme string, hljsTheme string) error {
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		return err
	}
	inFile, err := os.Open(inputMd)
	if err != nil {
		return err
	}
	defer inFile.Close()
	slides, err := split(io.Reader(inFile))
	if err != nil {
		return err
	}
	for idx, slide := range slides {
		slideNum := idx + 1
		outFile := filepath.Join(outDir, "content-"+strconv.Itoa(slideNum)+".html")
		renderedSlide, err := renderSlide(slide, false, slideNum, len(slides), theme, hljsTheme)
		if err != nil {
			return err
		}
		err = os.WriteFile(outFile, []byte(renderedSlide), 0755)
		if err != nil {
			return err
		}
	}
	srcDir := filepath.Join(filepath.Dir(inputMd), "assets")
	targetDir := filepath.Join(outDir, "assets")
	err = copyAssets(srcDir, targetDir)
	if err != nil {
		return fmt.Errorf("failed to copy assets: %v", err)
	}
	err = copyLibs(outDir)
	if err != nil {
		return fmt.Errorf("failed to copy libs: %v", err)
	}
	return nil
}

type RenderParams struct {
	IsPresenter  bool
	CurrentSlide int
	TotalSlides  int
	NextSlide    int
	PrevSlide    int
	Slide        Slide
	Theme        string
	HljsTheme    string
}

func toRenderParams(slide Slide, isPresenter bool, currentSlide int, totalSlides int, theme string, hljsTheme string) RenderParams {
	renderParams := RenderParams{
		IsPresenter:  isPresenter,
		CurrentSlide: currentSlide,
		TotalSlides:  totalSlides,
		Slide:        slide,
		Theme:        theme,
		HljsTheme:    hljsTheme,
	}
	if currentSlide < 2 {
		renderParams.PrevSlide = 1
	} else {
		renderParams.PrevSlide = currentSlide - 1
	}

	if currentSlide >= totalSlides {
		renderParams.NextSlide = totalSlides
	} else {
		renderParams.NextSlide = currentSlide + 1
	}

	return renderParams
}

func renderSlide(slide Slide, isPresenter bool, currentSlide int, totalSlides int, theme string, hljsTheme string) (string, error) {
	renderParams := toRenderParams(slide, isPresenter, currentSlide, totalSlides, theme, hljsTheme)
	tpl := template.Must(template.New("slide").Parse(string(slideHtml)))
	var buff bytes.Buffer
	err := tpl.Execute(&buff, renderParams)
	return buff.String(), err
}

func copyAssets(srcDir string, targetDir string) error {
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		return err
	}
	return os.CopyFS(targetDir, os.DirFS(srcDir))
}

func copyLibs(targetDir string) error {
	return os.CopyFS(targetDir, templateFs)
}

func renderHTML(mergedLines string) string {
	extensions := parser.CommonExtensions |
		parser.AutoHeadingIDs

	p := parser.NewWithExtensions(extensions)
	renderer := mdhtml.NewRenderer(mdhtml.RendererOptions{
		Flags: mdhtml.CommonFlags,
	})
	res := string(markdown.ToHTML([]byte(mergedLines), p, renderer))
	return res
}

func split(md io.Reader) ([]Slide, error) {
	h1 := regexp.MustCompile("^[\\s]*#\\s.*$")
	h2 := regexp.MustCompile("^[\\s]*##\\s.*$")
	startNotes := regexp.MustCompile("^:notes-start")
	endNotes := regexp.MustCompile("^:notes-end")
	mdScanner := bufio.NewScanner(md)
	inNotesSection := false
	// ignore leading blank lines
	onSlide := false
	var slide Slide
	var slides []Slide
	lineNum := 0

	for mdScanner.Scan() {
		curLine := mdScanner.Bytes()
		lineNum += 1
		switch {
		case h1.Match(curLine):
			// handle special case where multiple presentations are combined in one md file
			if len(slide.Title) > 0 {
				slides = append(slides, slide)
			}
			onSlide = true
			normalizedTitle := strings.ReplaceAll(string(curLine), "#", "")
			slide = Slide{
				IsFrontPage: true,
				Title:       normalizedTitle,
			}
		case h2.Match(curLine):
			slides = append(slides, slide)
			onSlide = true
			normalizedTitle := strings.ReplaceAll(string(curLine), "#", "")
			slide = Slide{
				IsFrontPage: false,
				Title:       normalizedTitle,
			}
		case startNotes.Match(curLine):
			if inNotesSection {
				return nil, fmt.Errorf("error on line %d: Nested notes are not allowed", lineNum)
			}
			inNotesSection = true
			continue
		case endNotes.Match(curLine):
			if !inNotesSection {
				return nil, fmt.Errorf("error on line %d: Superfluous :end-notes", lineNum)
			}
			inNotesSection = false
			continue
		case inNotesSection:
			slide.Notes = append(slide.Notes, string(curLine))
			continue
		case onSlide:
			slide.Content = append(slide.Content, string(curLine))
		}
	}
	slides = append(slides, slide)
	return slides, nil
}
