package ops

import "testing"

func TestIsBot(t *testing.T) {
	tests := []struct {
		author string
		want   bool
	}{
		{"dependabot[bot]", true},
		{"renovate[bot]", true},
		{"some-bot", true},
		{"dependabot", true},
		{"renovate", true},
		{"alice", false},
		{"bob", false},
		{"botman", false},   // doesn't end with [bot] or -bot
		{"[bot]user", false}, // suffix check, not prefix
		{"my-robot", false},  // doesn't end with -bot
	}
	for _, tt := range tests {
		t.Run(tt.author, func(t *testing.T) {
			if got := IsBot(tt.author); got != tt.want {
				t.Errorf("IsBot(%q) = %v, want %v", tt.author, got, tt.want)
			}
		})
	}
}

func TestMatchesAnyPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prefixes []string
		want     bool
	}{
		{"matches first", "web-app-foo", []string{"web-app-", "Camunda-"}, true},
		{"matches second", "Camunda-thing", []string{"web-app-", "Camunda-"}, true},
		{"no match", "my-service", []string{"web-app-", "Camunda-"}, false},
		{"empty prefixes", "anything", nil, false},
		{"empty name", "", []string{"web-app-"}, false},
		{"exact prefix", "web-app-", []string{"web-app-"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesAnyPrefix(tt.input, tt.prefixes); got != tt.want {
				t.Errorf("matchesAnyPrefix(%q, %v) = %v, want %v", tt.input, tt.prefixes, got, tt.want)
			}
		})
	}
}
