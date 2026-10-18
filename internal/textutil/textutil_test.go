package textutil

import (
	"reflect"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Hello, World!":          "hello-world",
		"  Crème brûlée  recipe": "creme-brulee-recipe",
		"Go 1.26 — what's new?":  "go-1-26-what-s-new",
		"---":                    "",
		"日本語 タイトル":               "日本語-タイトル",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseTags(t *testing.T) {
	got := ParseTags(" Go, web  dev ,go,, GO ,Testing", 5)
	want := []string{"Go", "web dev", "Testing"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseTags = %q, want %q", got, want)
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("short", 10); got != "short" {
		t.Fatalf("got %q", got)
	}
	if got := Truncate("the quick brown fox jumps", 12); got != "the quick…" {
		t.Fatalf("got %q", got)
	}
}

func TestReadingMinutes(t *testing.T) {
	if ReadingMinutes("") != 1 {
		t.Fatal("empty text should be 1 minute")
	}
	words := make([]byte, 0, 2*440)
	for range 440 {
		words = append(words, 'a', ' ')
	}
	if got := ReadingMinutes(string(words)); got != 2 {
		t.Fatalf("440 words = %d minutes, want 2", got)
	}
}
