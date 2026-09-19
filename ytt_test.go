package main

import (
	"strings"
	"testing"
)

func TestExtractVideoID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "full watch URL", input: "https://www.youtube.com/watch?v=S8INv2Q9GbA", want: "S8INv2Q9GbA"},
		{name: "short URL", input: "https://youtu.be/S8INv2Q9GbA?t=30", want: "S8INv2Q9GbA"},
		{name: "shorts URL", input: "https://www.youtube.com/shorts/S8INv2Q9GbA", want: "S8INv2Q9GbA"},
		{name: "embed URL", input: "https://www.youtube.com/embed/S8INv2Q9GbA", want: "S8INv2Q9GbA"},
		{name: "raw ID", input: "S8INv2Q9GbA", want: "S8INv2Q9GbA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractVideoID(tt.input)
			if err != nil {
				t.Fatalf("extractVideoID(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("extractVideoID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractVideoIDRejectsInvalidInput(t *testing.T) {
	_, err := extractVideoID("https://example.com/not-a-youtube-video")
	if err == nil {
		t.Fatal("extractVideoID accepted invalid input")
	}
	if want := "could not extract video ID"; !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want it to contain %q", err, want)
	}
}
