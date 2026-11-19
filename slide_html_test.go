package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSplitInvalidNotes(t *testing.T) {
	tests := []struct {
		name string
		md   string
	}{
		{
			name: "nested notes",
			md:   "# Title\n:notes-start\n:notes-start\n",
		},
		{
			name: "end without start",
			md:   "# Title\n:notes-end\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := split(strings.NewReader(tt.md))
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestRenderHTML(t *testing.T) {
	html := renderHTML("# Heading\n\nSome **bold** text.\n")

	for _, expected := range []string{
		"<h1",
		"Heading",
		"<strong>bold</strong>",
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML does not contain %q: %s", expected, html)
		}
	}
}

func TestRenderSlide(t *testing.T) {
	slide := Slide{
		IsFrontPage: false,
		Title:       "Test Slide",
		Content:     []string{"Some slide content."},
		Notes:       []string{"Internal notes."},
	}

	tests := []struct {
		name        string
		presenter   bool
		notesHidden bool
	}{
		{
			name:        "presenter view",
			presenter:   true,
			notesHidden: false,
		},
		{
			name:        "viewer view",
			presenter:   false,
			notesHidden: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := renderSlide(
				slide, tt.presenter, 1, 2, "theme.css", "highlight.css",
			)
			if err != nil {
				t.Fatalf("renderSlide failed: %v", err)
			}

			for _, expected := range []string{
				"Test Slide",
				"Some slide content.",
				"Internal notes.",
				"1/2",
			} {
				if !strings.Contains(html, expected) {
					t.Errorf("expected HTML to contain %q", expected)
				}
			}

			if tt.notesHidden && !strings.Contains(html, "hidden") {
				t.Error("viewer notes should be hidden")
			}
		})
	}
}

func TestToRenderParams(t *testing.T) {
	tests := []struct {
		name                       string
		current, total             int
		expectedPrev, expectedNext int
	}{
		{"first slide", 1, 3, 1, 2},
		{"middle slide", 2, 3, 1, 3},
		{"last slide", 3, 3, 2, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := toRenderParams(
				Slide{}, false, tt.current, tt.total, "light", "light",
			)

			if params.PrevSlide != tt.expectedPrev {
				t.Errorf("PrevSlide = %d, want %d",
					params.PrevSlide, tt.expectedPrev)
			}

			if params.NextSlide != tt.expectedNext {
				t.Errorf("NextSlide = %d, want %d",
					params.NextSlide, tt.expectedNext)
			}
		})
	}
}

func TestSlideContentAndNotesAsHTML(t *testing.T) {
	slide := Slide{
		Content: []string{"A **bold** sentence."},
		Notes:   []string{"A *private* note."},
	}

	content := slide.ContentAsHtml()
	notes := slide.NotesAsHtml()

	if !strings.Contains(content, "<strong>bold</strong>") {
		t.Errorf("unexpected content HTML: %s", content)
	}

	if !strings.Contains(notes, "<em>private</em>") {
		t.Errorf("unexpected notes HTML: %s", notes)
	}
}

// Keep bytes imported for compatibility with the existing
// Markdown parsing test conventions.
func TestSplitBasic(t *testing.T) {
	md := "# Presentation\nSubtitle\n## First Slide\nHello\n"

	slides, err := split(bytes.NewBufferString(md))
	if err != nil {
		t.Fatalf("split failed: %v", err)
	}

	if len(slides) != 2 {
		t.Fatalf("got %d slides, want 2", len(slides))
	}

	if !slides[0].IsFrontPage {
		t.Error("first slide should be the front page")
	}

	if slides[1].IsFrontPage {
		t.Error("second slide should not be the front page")
	}
}

func TestExamplePresentation(t *testing.T) {
	slides, err := func() ([]Slide, error) {
		file, err := os.Open("example/example.md")
		if err != nil {
			return nil, err
		}
		defer file.Close()

		return split(file)
	}()

	if err != nil {
		t.Fatalf("could not parse example presentation: %v", err)
	}

	if len(slides) < 2 {
		t.Fatalf("expected multiple slides, got %d", len(slides))
	}

	for i, slide := range slides {
		html, err := renderSlide(
			slide, true, i+1, len(slides), "light.css", "light-hl.css",
		)
		if err != nil {
			t.Fatalf("could not render slide %d: %v", i+1, err)
		}

		if !strings.Contains(html, "<html") {
			t.Errorf("slide %d does not contain HTML document", i+1)
		}
	}
}
