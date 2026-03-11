package ui

import "testing"

func TestParseCommaSeparated(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", []string{}},
		{"spaces only", "  ", []string{}},
		{"single value", "foo", []string{"foo"}},
		{"two values", "foo, bar", []string{"foo", "bar"}},
		{"three values", "foo, bar, baz", []string{"foo", "bar", "baz"}},
		{"extra spaces", " foo , bar , ", []string{"foo", "bar"}},
		{"empty segments", ",,foo,,", []string{"foo"}},
		{"no spaces", "a,b,c", []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommaSeparated(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseCommaSeparated(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseCommaSeparated(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}
