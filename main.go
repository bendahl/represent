package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

//go:embed assets/highlight.min.js
var highlightJS string

//go:embed assets/highlight.min.css
var highlightDarkCSS string

//go:embed assets/highlight-light.min.css
var highlightLightCSS string

type Slide struct {
	Title    string
	Subtitle string
	Markdown string
	Notes    string
	IsTitle  bool
}

var (
	titleRE = regexp.MustCompile(`^#\s+(.+?)\s*$`)
	slideRE = regexp.MustCompile(`^##\s+(.+?)\s*$`)
)

type PresentationServer struct {
	mu      sync.Mutex
	current int // zero-based slide index
	total   int
	clients map[chan int]struct{}
	dir     string
}

func newPresentationServer(dir string) (*PresentationServer, error) {
	files, err := filepath.Glob(filepath.Join(dir, "slide-*.html"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no slide-*.html files found in %s", dir)
	}

	return &PresentationServer{
		total:   len(files),
		clients: make(map[chan int]struct{}),
		dir:     dir,
	}, nil
}

func (s *PresentationServer) index() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

func (s *PresentationServer) broadcast() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for ch := range s.clients {
		// Keep only the latest state for slow clients.
		select {
		case ch <- s.current:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- s.current:
			default:
			}
		}
	}
}

func (s *PresentationServer) move(delta int) {
	s.mu.Lock()
	next := s.current + delta

	if next < 0 {
		next = 0
	}
	if next >= s.total {
		next = s.total - 1
	}

	changed := next != s.current
	s.current = next
	s.mu.Unlock()

	if changed {
		s.broadcast()
	}
}

func (s *PresentationServer) events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan int, 1)

	s.mu.Lock()
	s.clients[ch] = struct{}{}
	initial := s.current
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
	}()

	fmt.Fprintf(w, "data: %d\n\n", initial)
	flusher.Flush()

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case index := <-ch:
			fmt.Fprintf(w, "data: %d\n\n", index)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func (s *PresentationServer) page(presenter bool) string {
	label := "Slides"

	if presenter {
		label = "Presenter"
	}

	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<style>
html, body {
  margin: 0;
  width: 100%%;
  height: 100%%;
  overflow: hidden;
  background: #111827;
}
iframe {
  display: block;
  width: 100%%;
  height: 100%%;
  border: 0;
}
</style>
</head>
<body>
<iframe id="view" title="%s"></iframe>
<script>
(() => {
  const view = document.getElementById("view");
  const prefix = %q;
  let current = -1;

  function show(index) {
    index = Number(index);
    if (!Number.isInteger(index) || index < 0 || index >= %d) return;
    if (index === current) return;
    current = index;
    view.src = "/file/" + prefix + String(index + 1).padStart(2, "0") + ".html";
  }

  async function move(direction) {
    try {
      await fetch(direction > 0 ? "/next" : "/prev", { method: "POST" });
    } catch (e) {
      console.error("Could not change slide", e);
    }
  }

  // Listen in this window and inside same-origin slide documents.
  function bindKeys(win) {
    try {
      if (win.__slideKeysBound) return;
      win.__slideKeysBound = true;

      win.addEventListener("keydown", event => {
        if (event.altKey || event.ctrlKey || event.metaKey) return;
        const target = event.target;
        if (target && (
          target.isContentEditable ||
          /^(INPUT|TEXTAREA|SELECT)$/.test(target.tagName)
        )) return;

        if (event.key === "ArrowRight" || event.key === "PageDown") {
          event.preventDefault();
          move(1);
        } else if (event.key === "ArrowLeft" || event.key === "PageUp") {
          event.preventDefault();
          move(-1);
        }
      });

      for (const frame of win.document.querySelectorAll("iframe")) {
        frame.addEventListener("load", () => bindKeys(frame.contentWindow));
        if (frame.contentDocument && frame.contentDocument.readyState === "complete") {
          bindKeys(frame.contentWindow);
        }
      }
    } catch (_) {
      // Ignore inaccessible frames.
    }
  }

  view.addEventListener("load", () => bindKeys(view.contentWindow));
  bindKeys(window);

  function connect() {
    const events = new EventSource("/events");
    events.onmessage = event => show(event.data);
    events.onerror = () => {
      events.close();
      setTimeout(connect, 1500);
    };
  }

  connect();
})();
</script>
</body>
</html>`, label, label, prefixFor(presenter), s.total)
}

func prefixFor(presenter bool) string {
	if presenter {
		return "presenter-"
	}
	return "slide-"
}

func (s *PresentationServer) serve() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/slides", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/slides/", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/slides/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, s.page(false))
	})

	mux.HandleFunc("/present", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/present/", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/present/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, s.page(true))
	})

	mux.HandleFunc("/events", s.events)

	mux.HandleFunc("/next", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		s.move(1)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/prev", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		s.move(-1)
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/file/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/file/")
		if name == "" || filepath.Base(name) != name ||
			(!strings.HasPrefix(name, "slide-") &&
				!strings.HasPrefix(name, "presenter-")) ||
			!strings.HasSuffix(name, ".html") {
			http.NotFound(w, r)
			return
		}

		data, err := os.ReadFile(filepath.Join(s.dir, name))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Write(data)
	})

	fmt.Println("Slides:    http://localhost:8080/slides")
	fmt.Println("Presenter: http://localhost:8080/present")
	return http.ListenAndServe("localhost:8080", mux)
}

func isFenceLine(s string) bool {
	return strings.HasPrefix(s, "```") ||
		strings.HasPrefix(s, "~~~")
}

// splitSlides separates slide content from presenter notes.
// Headings and note markers inside fenced code blocks are ignored.
func splitSlides(source string) ([]Slide, error) {
	scanner := bufio.NewScanner(strings.NewReader(source))
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	var slides []Slide
	var subtitle []string
	var notes []string

	foundTitle := false
	inRegularSlide := false
	inNotes := false

	inFence := false
	var fenceChar byte
	fenceLen := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		wasInFence := inFence

		if !inFence && isFenceLine(trimmed) {
			fenceChar = trimmed[0]
			fenceLen = 0
			for fenceLen < len(trimmed) &&
				trimmed[fenceLen] == fenceChar {
				fenceLen++
			}
			if fenceLen >= 3 {
				inFence = true
			}
		} else if inFence {
			n := 0
			for n < len(trimmed) && trimmed[n] == fenceChar {
				n++
			}
			if n >= fenceLen && strings.TrimSpace(trimmed[n:]) == "" {
				inFence = false
			}
		}

		// Structural markers are only meaningful outside code fences.
		outsideFence := !wasInFence && !isFenceLine(trimmed)

		if outsideFence && !foundTitle {
			if m := titleRE.FindStringSubmatch(line); m != nil {
				foundTitle = true
				slides = append(slides, Slide{
					Title:   m[1],
					IsTitle: true,
				})
				continue
			}
		}

		if !foundTitle {
			continue
		}

		if outsideFence {
			if trimmed == ":notes-start" {
				if inNotes {
					return nil, fmt.Errorf("nested :notes-start marker")
				}
				inNotes = true
				continue
			}

			if trimmed == ":notes-end" {
				if !inNotes {
					return nil, fmt.Errorf(":notes-end without :notes-start")
				}
				inNotes = false
				continue
			}

			// Every h2 begins a new slide, never a new slide inside notes.
			if !inNotes {
				if m := slideRE.FindStringSubmatch(line); m != nil {
					if !inRegularSlide {
						slides[0].Subtitle = normalizeSubtitle(subtitle)
						subtitle = nil
					}
					inRegularSlide = true
					slides = append(slides, Slide{
						Markdown: line,
					})
					continue
				}
			}
		}

		current := &slides[len(slides)-1]

		if inNotes {
			notes = append(notes, line)
		} else if !inRegularSlide {
			subtitle = append(subtitle, line)
		} else {
			current.Markdown += "\n" + line
		}

		// Notes are accumulated separately for the current slide.
		if inNotes {
			current.Notes = strings.Join(notes, "\n")
		} else {
			notes = nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading Markdown: %w", err)
	}
	if !foundTitle {
		return nil, fmt.Errorf("no title found; start with '# Title'")
	}
	if inNotes {
		return nil, fmt.Errorf("missing :notes-end marker")
	}
	if !inRegularSlide {
		slides[0].Subtitle = normalizeSubtitle(subtitle)
	}

	return slides, nil
}

func normalizeSubtitle(lines []string) string {
	return strings.Join(strings.Fields(strings.Join(lines, " ")), " ")
}

func renderMarkdown(source string) string {
	extensions := parser.CommonExtensions |
		parser.Tables |
		parser.FencedCode |
		parser.Strikethrough |
		parser.AutoHeadingIDs

	p := parser.NewWithExtensions(extensions)
	renderer := mdhtml.NewRenderer(mdhtml.RendererOptions{
		Flags: mdhtml.CommonFlags,
	})

	return string(markdown.ToHTML([]byte(source), p, renderer))
}

func baseCSS(theme string) string {
	colors := `
:root {
  --bg: #111827;
  --fg: #e5e7eb;
  --muted: #9ca3af;
  --border: #4b5563;
  --surface: #1f2937;
  --inline-code: #374151;
}
`
	if theme == "light" {
		colors = `
:root {
  --bg: #ffffff;
  --fg: #1f2937;
  --muted: #6b7280;
  --border: #d1d5db;
  --surface: #f3f4f6;
  --inline-code: #e5e7eb;
}
`
	}

	return colors + `
* { box-sizing: border-box; }

body {
  margin: 0;
  padding: 3rem 6vw 5rem;
  font-family: system-ui, -apple-system, sans-serif;
  font-size: clamp(1rem, 2vw, 1.5rem);
  line-height: 1.55;
  color: var(--fg);
  background: var(--bg);
}

main { max-width: 1100px; margin: 0 auto; }
h1, h2, h3 { line-height: 1.2; }
h1 { font-size: clamp(2.5rem, 6vw, 5rem); }
h2 { font-size: clamp(2rem, 4vw, 3rem); }
h3 {
  font-size: clamp(1.3rem, 2.5vw, 2rem);
  color: var(--muted);
}
.title-slide {
  min-height: 75vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  text-align: center;
}
.title-slide h1 { margin-bottom: 1rem; }
.title-slide h3 { font-weight: 400; }
a { color: #60a5fa; }
img { max-width: 100%; height: auto; }
table { border-collapse: collapse; width: 100%; margin: 1.5rem 0; }
th, td {
  border: 1px solid var(--border);
  padding: .6rem .8rem;
  text-align: left;
}
th { background: var(--surface); }
pre {
  overflow-x: auto;
  padding: 1.2rem;
  border-radius: .5rem;
  background: var(--surface);
  font-size: .85em;
}
code {
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
}
:not(pre) > code {
  background: var(--inline-code);
  padding: .15em .35em;
  border-radius: .25rem;
}
blockquote {
  border-left: 4px solid var(--border);
  padding-left: 1rem;
  color: var(--muted);
}
.slide-number {
  position: fixed;
  right: 1rem;
  bottom: .75rem;
  z-index: 1000;
  padding: .35rem .65rem;
  border: 1px solid var(--border);
  border-radius: .4rem;
  background: var(--bg);
  color: var(--muted);
  font-size: .8rem;
}
`
}

func highlightCSS(theme string) string {
	if theme == "light" {
		return highlightLightCSS
	}
	return highlightDarkCSS
}

func writeAssetsStyle(b *bytes.Buffer, theme string) {
	b.WriteString("<style>\n")
	b.WriteString(baseCSS(theme))
	b.WriteString("\n")
	b.WriteString(highlightCSS(theme))
	b.WriteString("\n</style>\n")
}

func writeHighlightJS(b *bytes.Buffer) {
	b.WriteString("<script>\n")
	b.WriteString(highlightJS)
	b.WriteString("\n</script>\n")
	b.WriteString(`<script>
if (window.hljs) {
  document.querySelectorAll('pre code').forEach(function (block) {
    hljs.highlightElement(block);
  });
}
</script>`)
}

func renderSlide(slide Slide, number, total int, theme string) string {
	var b bytes.Buffer

	b.WriteString(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>`)

	if slide.IsTitle {
		b.WriteString(html.EscapeString(slide.Title))
	} else {
		b.WriteString(fmt.Sprintf("Slide %d", number))
	}

	b.WriteString("</title>\n")
	writeAssetsStyle(&b, theme)

	b.WriteString("</head>\n<body>\n<main>\n")

	if slide.IsTitle {
		b.WriteString(`<section class="title-slide"><h1>`)
		b.WriteString(html.EscapeString(slide.Title))
		b.WriteString("</h1>")
		if slide.Subtitle != "" {
			b.WriteString("<h3>")
			b.WriteString(html.EscapeString(slide.Subtitle))
			b.WriteString("</h3>")
		}
		b.WriteString("</section>")
	} else {
		b.WriteString(renderMarkdown(slide.Markdown))
	}

	b.WriteString("</main>\n")
	b.WriteString(fmt.Sprintf(
		`<div class="slide-number">%d / %d</div>`,
		number, total,
	))
	writeHighlightJS(&b)
	b.WriteString("\n</body>\n</html>\n")

	return b.String()
}

func renderPresenter(number, total int, notes, theme string) string {
	var b bytes.Buffer

	b.WriteString(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Presenter - Slide `)
	b.WriteString(fmt.Sprint(number))
	b.WriteString("</title>\n")

	b.WriteString("<style>\n")
	b.WriteString(baseCSS(theme))
	b.WriteString(`
body {
  padding: 0;
  height: 100vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.slide-frame {
  width: 100%;
  height: 68vh;
  min-height: 0;
  border: 0;
  border-bottom: 2px solid var(--border);
  background: var(--bg);
}
.notes {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 1rem 6vw 3rem;
}
.notes h2 { font-size: 1.5rem; }
.notes h3 { font-size: 1.2rem; }
.notes pre { font-size: .8rem; }
.presenter-label {
  margin: 0 0 .5rem;
  color: var(--muted);
  font-size: .8rem;
  text-transform: uppercase;
  letter-spacing: .08em;
}
.slide-number { bottom: .5rem; right: .75rem; }
`)
	b.WriteString(highlightCSS(theme))
	b.WriteString("\n</style>\n</head>\n<body>\n")

	fmt.Fprintf(&b,
		`<iframe class="slide-frame" src="slide-%02d.html" title="Current slide"></iframe>`,
		number,
	)

	b.WriteString(`<section class="notes">
<div class="presenter-label">Presenter notes</div>
<div class="notes-content">`)

	//if strings.TrimSpace(notes) == "" {
	//	b.WriteString(`<p>No notes for this slide.</p>`)
	//} else {
	b.WriteString(renderMarkdown(notes))
	//}

	b.WriteString(`</div></section>`)

	fmt.Fprintf(&b,
		`<div class="slide-number">%d / %d</div>`,
		number, total,
	)

	// The presenter page has its own Highlight.js instance for notes.
	writeHighlightJS(&b)

	b.WriteString("\n</body>\n</html>\n")
	return b.String()
}

func main() {
	present := flag.Bool("present", false, "serve an existing presentation")
	outDir := flag.String("out", "slides", "output directory")
	theme := flag.String("theme", "dark", "color theme: light or dark")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: \nGenerate: %s [options] input.md\n\nOptions:\nPresent: %s -present slides-folder\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if *present {
		if flag.NArg() != 1 {
			fmt.Fprintf(os.Stderr,
				"Usage: %s -present <slides-folder>\n", os.Args[0])
			os.Exit(2)
		}

		dir, err := filepath.Abs(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve slides directory: %v\n", err)
			os.Exit(1)
		}

		server, err := newPresentationServer(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := server.serve(); err != nil {
			fmt.Fprintf(os.Stderr, "presentation server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	if *theme != "light" && *theme != "dark" {
		fmt.Fprintln(os.Stderr, "theme must be 'light' or 'dark'")
		os.Exit(2)
	}

	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		os.Exit(1)
	}

	slides, err := splitSlides(string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse slides: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "create output directory: %v\n", err)
		os.Exit(1)
	}

	total := len(slides)

	for i, slide := range slides {
		number := i + 1

		slidePath := filepath.Join(
			*outDir, fmt.Sprintf("slide-%02d.html", number),
		)
		presenterPath := filepath.Join(
			*outDir, fmt.Sprintf("presenter-%02d.html", number),
		)

		slideHTML := renderSlide(slide, number, total, *theme)
		presenterHTML := renderPresenter(
			number, total, slide.Notes, *theme,
		)

		if err := os.WriteFile(slidePath, []byte(slideHTML), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", slidePath, err)
			os.Exit(1)
		}
		if err := os.WriteFile(
			presenterPath, []byte(presenterHTML), 0644,
		); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", presenterPath, err)
			os.Exit(1)
		}

		fmt.Println("Created", slidePath)
		fmt.Println("Created", presenterPath)
	}
}
