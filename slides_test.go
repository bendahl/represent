package main

import (
	"bytes"
	"fmt"
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
	fmt.Printf("%#v", slides)
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
	fmt.Printf("%#v", slides)
}
