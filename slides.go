package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Slide struct {
	IsFrontPage bool
	Title       string
	Content     []string
	Notes       []string
}

func (s Slide) ContentAsMD() string {
	mergedLines := strings.Join(s.Content, "\n")
	return renderHTML(mergedLines)
}

func (s Slide) NotesAsMD() string {
	mergedLines := strings.Join(s.Notes, "\n")
	return renderHTML(mergedLines)
}

func renderHTML(mergedLines string) string {
	extensions := parser.CommonExtensions |
		parser.Tables |
		parser.FencedCode |
		parser.Strikethrough |
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
	startNotes := regexp.MustCompile("^:start-notes")
	endNotes := regexp.MustCompile("^:end-notes")
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
			slide = Slide{
				IsFrontPage: true,
				Title:       string(curLine),
			}
		case h2.Match(curLine):
			slides = append(slides, slide)
			onSlide = true
			slide = Slide{
				IsFrontPage: false,
				Title:       string(curLine),
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
