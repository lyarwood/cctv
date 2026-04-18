package tui

import (
	"testing"

	"github.com/lyarwood/cctv/internal/claude"
)

func TestMatchSession(t *testing.T) {
	s := claude.Session{
		Summary:     "Writing unit tests in Go",
		FirstPrompt: "How do I write tests?",
		ProjectPath: "/home/user/myproject",
		GitBranch:   "feature/auth",
		PRLinks: []claude.PRLink{
			{Number: 42, URL: "https://github.com/user/myproject/pull/42", Repository: "user/myproject"},
		},
	}

	tests := []struct {
		name  string
		text  string
		match bool
	}{
		{"bare text matches summary", "unit tests", true},
		{"bare text matches first prompt", "write tests", true},
		{"bare text matches project path", "myproject", true},
		{"bare text matches branch", "auth", true},
		{"bare text no match", "nonexistent", false},
		{"bare text case insensitive", "UNIT TESTS", true},

		{"project prefix matches basename", "project:myproject", true},
		{"project prefix matches full path", "project:/home/user", true},
		{"project prefix no match", "project:otherproject", false},
		{"project prefix case insensitive", "project:MYPROJECT", true},

		{"branch prefix matches", "branch:feature/auth", true},
		{"branch prefix partial match", "branch:auth", true},
		{"branch prefix no match", "branch:main", false},
		{"branch prefix case insensitive", "branch:FEATURE", true},

		{"cwd prefix matches path", "cwd:/home/user", true},
		{"cwd prefix partial match", "cwd:myproject", true},
		{"cwd prefix no match", "cwd:/other/path", false},
		{"cwd prefix case insensitive", "cwd:HOME", true},

		{"empty filter matches", "", true},

		{"unknown prefix falls through to bare search", "foo:bar", false},
		{"unknown prefix matching content", "Go:whatever", false},

		{"pr prefix matches repo", "pr:user/myproject", true},
		{"pr prefix matches number", "pr:42", true},
		{"pr prefix matches repo#number", "pr:myproject#42", true},
		{"pr prefix no match", "pr:99", false},
		{"pr prefix case insensitive", "pr:USER/MYPROJECT", true},

		{"combined project and branch", "project:myproject branch:auth", true},
		{"combined project and branch mismatch", "project:myproject branch:main", false},
		{"combined bare and prefix", "unit project:myproject", true},
		{"combined bare and prefix mismatch", "unit project:other", false},
		{"multiple bare terms", "unit Go", true},
		{"multiple bare terms one missing", "unit Python", false},

		{"regex anchor end", "project:myproject$", true},
		{"regex anchor end excludes partial", "project:myprojec$", false},
		{"regex anchor start", "project:^myproject", true},
		{"regex dot star", "project:my.*ject", true},
		{"regex bare matches field end", "tests\\?$", true},
		{"regex bare no match", "^Python", false},
		{"regex branch anchor", "branch:auth$", true},
		{"regex branch excludes", "branch:^auth$", false},
		{"regex pr number", "pr:#42$", true},
		{"invalid regex falls back to substring", "project:myproject", true},
		{"invalid regex bracket literal", "project:my[", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.text == "" {
				// Empty always matches in the caller (applyFilter skips matchSession)
				return
			}
			got := matchSession(s, tt.text)
			if got != tt.match {
				t.Errorf("matchSession(%q) = %v, want %v", tt.text, got, tt.match)
			}
		})
	}
}

func TestMatchSessions(t *testing.T) {
	sessions := []claude.Session{
		{Summary: "Auth work", ProjectPath: "/home/user/auth", GitBranch: "main"},
		{Summary: "API work", ProjectPath: "/home/user/api", GitBranch: "develop"},
		{Summary: "UI work", ProjectPath: "/home/user/frontend", GitBranch: "main"},
	}

	t.Run("project filter narrows results", func(t *testing.T) {
		got := matchSessions(sessions, "project:auth")
		if len(got) != 1 || got[0].Summary != "Auth work" {
			t.Errorf("expected 1 result for project:auth, got %d", len(got))
		}
	})

	t.Run("branch filter narrows results", func(t *testing.T) {
		got := matchSessions(sessions, "branch:main")
		if len(got) != 2 {
			t.Errorf("expected 2 results for branch:main, got %d", len(got))
		}
	})

	t.Run("bare text searches all fields", func(t *testing.T) {
		got := matchSessions(sessions, "work")
		if len(got) != 3 {
			t.Errorf("expected 3 results for 'work', got %d", len(got))
		}
	})

	t.Run("no matches returns empty", func(t *testing.T) {
		got := matchSessions(sessions, "branch:release")
		if len(got) != 0 {
			t.Errorf("expected 0 results for branch:release, got %d", len(got))
		}
	})

	t.Run("combined filters narrow results", func(t *testing.T) {
		got := matchSessions(sessions, "project:auth branch:main")
		if len(got) != 1 || got[0].Summary != "Auth work" {
			t.Errorf("expected 1 result for project:auth branch:main, got %d", len(got))
		}
	})

	t.Run("combined filters with no overlap", func(t *testing.T) {
		got := matchSessions(sessions, "project:api branch:main")
		if len(got) != 0 {
			t.Errorf("expected 0 results for project:api branch:main, got %d", len(got))
		}
	})
}
