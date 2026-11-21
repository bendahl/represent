package main

import (
	"bytes"
	"strings"
	"testing"
)

const mdFile = `
# This is a Test Front Page
This is a subtitle

:notes-start
- These are presenter notes
- Another note
	- Some sub list
:notes-end

## First Regular slide
Some Test
## Second Slide
More Text
~~~go
package main
func main() {
	return
}
~~~
:notes-start
Some note
:notes-end
 ## Header with Space in Front should work as well...
`

func TestSplit(t *testing.T) {
	slides, err := split(bytes.NewReader([]byte(mdFile)))
	if err != nil {
		t.Fatalf("Unexpected error during split operation: %v", err)
	}
	numSlidesExpected := 4
	if len(slides) != numSlidesExpected {
		t.Fatalf("Expected %d slides. Actual: %d", numSlidesExpected, len(slides))
	}
}

func TestRender(t *testing.T) {
	expectedNotes := `<ul>
<li>These are presenter notes</li>
<li>Another note

<ul>
<li>Some sub list</li>
</ul></li>
</ul>
`
	slides, err := split(bytes.NewReader([]byte(mdFile)))
	if err != nil {
		t.Fatalf("Unexpected error during split operation: %v", err)
	}
	actualNotes := slides[0].NotesAsHtml()
	if actualNotes != expectedNotes {
		t.Fatalf("Notes format does not match expected format.")
	}

	expectedContent := "<p>This is a subtitle</p>\n"
	actualContent := slides[0].ContentAsHtml()
	if actualContent != expectedContent {
		t.Fatalf("Content format does not match expected format.")
	}
}

const multiPrentationMd = `
# This is a Test Front Page
This is a subtitle

:notes-start
#  H1Headers are handled correctly
## H2 Headers are handled correctly
- These are presenter notes
- Another note
	- Some sub list
:notes-end

## First Regular slide
Some Test
# Another Presentation starts...
Another Subtitle
:notes-start
Some note
:notes-end
 ## Header with Space in Front should work as well...
`

func TestMultiplePresentationMd(t *testing.T) {
	slides, err := split(bytes.NewReader([]byte(multiPrentationMd)))
	if err != nil {
		t.Fatalf("Unexpected error during split operation: %v", err)
	}
	numSlidesExpected := 4
	if len(slides) != numSlidesExpected {
		t.Fatalf("Expected %d slides. Actual: %d", numSlidesExpected, len(slides))
	}
}

func TestSplitCodeBlocks(t *testing.T) {
	tests := []struct {
		name           string
		md             string
		expectedSlides int
	}{
		{
			name: "HTML code block",
			md: "# Presentation\n\n" +
				"## First Slide\n\n" +
				"Some content.\n\n" +
				"<pre><code>\n" +
				"# This is not a slide\n" +
				"## Neither is this\n" +
				"</code></pre>\n\n" +
				"## Second Slide\n\n" +
				"More content.\n",
			expectedSlides: 3,
		},
		{
			name: "Markdown code block with backticks",
			md: "# Presentation\n\n" +
				"## First Slide\n\n" +
				"Some content.\n\n" +
				"````markdown\n" +
				"# This is not a slide\n" +
				"## Neither is this\n" +
				"````\n\n" +
				"## Second Slide\n\n" +
				"More content.\n",
			expectedSlides: 3,
		},
		{
			name: "Markdown code block with tildes",
			md: "# Presentation\n\n" +
				"## First Slide\n\n" +
				"Some content.\n\n" +
				"~~~markdown\n" +
				"# This is not a slide\n" +
				"## Neither is this\n" +
				"~~~\n\n" +
				"## Second Slide\n\n" +
				"More content.\n",
			expectedSlides: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slides, err := split(strings.NewReader(tt.md))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(slides) != tt.expectedSlides {
				t.Fatalf(
					"expected %d slides, got %d",
					tt.expectedSlides,
					len(slides),
				)
			}

			expectedTitles := []string{
				"Presentation",
				"First Slide",
				"Second Slide",
			}

			for i, expected := range expectedTitles {
				if slides[i].Title != expected {
					t.Errorf(
						"slide %d: expected title %q, got %q",
						i,
						expected,
						slides[i].Title,
					)
				}
			}
		})
	}
}
