package vector

import (
	"testing"
	"time"
)

func TestParseUTC_MatchesTimeParse(t *testing.T) {
	cases := []string{
		"2026-03-11T18:45:53Z",
		"2026-01-01T00:00:00Z",
		"2026-12-31T23:59:59Z",
		"2026-03-11T14:58:35Z",
	}
	for _, c := range cases {
		got, ok := parseUTC(c)
		if !ok {
			t.Errorf("parseUTC(%q) falhou", c)
			continue
		}
		want, err := time.Parse(time.RFC3339, c)
		if err != nil {
			t.Fatalf("time.Parse(%q): %v", c, err)
		}
		if !got.Equal(want) {
			t.Errorf("parseUTC(%q) = %v, want %v", c, got, want)
		}
	}
}

func TestParseUTC_Rejects(t *testing.T) {
	bad := []string{
		"",
		"2026-03-11T18:45:53",
		"2026-03-11T18:45:53+0000",
		"abc",
		"2026-13-01T00:00:00Z",
		"2026-03-32T00:00:00Z",
		"2026-03-11T24:00:00Z",
	}
	for _, c := range bad {
		if _, ok := parseUTC(c); ok {
			t.Errorf("parseUTC(%q) deveria ter falhado", c)
		}
	}
}

func BenchmarkParseUTC(b *testing.B) {
	s := "2026-03-11T18:45:53Z"
	for i := 0; i < b.N; i++ {
		_, _ = parseUTC(s)
	}
}

func BenchmarkTimeParse(b *testing.B) {
	s := "2026-03-11T18:45:53Z"
	for i := 0; i < b.N; i++ {
		_, _ = time.Parse(time.RFC3339, s)
	}
}
