package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("could not create directory: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("could not write test file: %v", err)
	}
}

func TestGenerateSlides(t *testing.T) {
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")
	inputFile := filepath.Join(inputDir, "presentation.md")

	markdown := `# Test Presentation
Test Subtitle

## First Slide
Some **bold** content.

:notes-start
These are presenter notes.
:notes-end

## Second Slide
Some more content.
`

	writeTestFile(t, inputFile, markdown)

	// Create an example asset.
	assetContent := "test image content"
	assetFile := filepath.Join(inputDir, "assets", "test.txt")
	writeTestFile(t, assetFile, assetContent)

	err := generateSlides(outputDir, inputFile, "light", "light")
	if err != nil {
		t.Fatalf("generateSlides failed: %v", err)
	}

	// Define expected content for each slide.
	expectedContents := [][]string{
		{
			"Test Presentation",
			"Test Subtitle",
			"1/3",
		},
		{
			"First Slide",
			"<strong>bold</strong>",
			"These are presenter notes.",
			"2/3",
		},
		{
			"Second Slide",
			"Some more content.",
			"3/3",
		},
	}

	// Verify all generated slides.
	for i := 1; i <= 3; i++ {
		slideFile := filepath.Join(
			outputDir,
			fmt.Sprintf("content-%d.html", i),
		)

		content, err := os.ReadFile(slideFile)
		if err != nil {
			t.Fatalf("could not read generated slide %d: %v", i, err)
		}

		slideContent := string(content)

		if !strings.Contains(slideContent, "<html") {
			t.Errorf("slide %d does not contain an HTML document", i)
		}

		for _, expected := range expectedContents[i-1] {
			if !strings.Contains(slideContent, expected) {
				t.Errorf("slide %d should contain %q", i, expected)
			}
		}
	}

	// Verify that assets were copied.
	copiedAsset := filepath.Join(outputDir, "assets", "test.txt")
	copiedContent, err := os.ReadFile(copiedAsset)
	if err != nil {
		t.Fatalf("asset was not copied: %v", err)
	}

	if string(copiedContent) != assetContent {
		t.Errorf("copied asset = %q, want %q",
			string(copiedContent), assetContent)
	}

	// Verify that embedded libraries were copied.
	if _, err := os.Stat(filepath.Join(outputDir, "libs")); err != nil {
		t.Errorf("embedded libraries were not copied: %v", err)
	}
}

func TestGenerateSlidesWithoutAssets(t *testing.T) {
	tmpDir := t.TempDir()

	inputFile := filepath.Join(tmpDir, "presentation.md")
	outputDir := filepath.Join(tmpDir, "output")

	writeTestFile(t, inputFile, `# Test Presentation

## First Slide
Some content.
`)

	err := generateSlides(outputDir, inputFile, "light", "light")
	if err != nil {
		t.Fatalf("generateSlides should succeed without assets: %v", err)
	}

	slideFile := filepath.Join(outputDir, "content-1.html")
	if _, err := os.Stat(slideFile); err != nil {
		t.Errorf("expected generated slide: %v", err)
	}

	assetsDir := filepath.Join(outputDir, "assets")
	if _, err := os.Stat(assetsDir); !os.IsNotExist(err) {
		t.Error("assets directory should not exist when input has no assets")
	}
}

func TestGenerateSlidesMissingInput(t *testing.T) {
	tmpDir := t.TempDir()

	inputFile := filepath.Join(tmpDir, "missing.md")
	outputDir := filepath.Join(tmpDir, "output")

	err := generateSlides(outputDir, inputFile, "light", "light")

	if err == nil {
		t.Fatal("expected an error for missing input file")
	}
}

func TestGenerateSlidesInvalidMarkdown(t *testing.T) {
	tmpDir := t.TempDir()

	inputFile := filepath.Join(tmpDir, "presentation.md")
	outputDir := filepath.Join(tmpDir, "output")

	writeTestFile(t, inputFile, `# Test Presentation

## First Slide

:notes-start
Some notes.

:notes-start
Nested notes.
`)

	err := generateSlides(outputDir, inputFile, "light", "light")

	if err == nil {
		t.Fatal("expected an error for nested presenter notes")
	}

	if !strings.Contains(err.Error(), "Nested notes are not allowed") {
		t.Errorf("unexpected error: %v", err)
	}
}
