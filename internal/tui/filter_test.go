package tui

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lyarwood/cctv/internal/claude"
)

var _ = Describe("matchSession", func() {
	var s claude.Session

	BeforeEach(func() {
		s = claude.Session{
			Summary:     "Writing unit tests in Go",
			FirstPrompt: "How do I write tests?",
			ProjectPath: "/home/user/myproject",
			GitBranch:   "feature/auth",
			PRLinks: []claude.PRLink{
				{Number: 42, URL: "https://github.com/user/myproject/pull/42", Repository: "user/myproject"},
			},
		}
	})

	DescribeTable("bare text",
		func(text string, expected bool) {
			Expect(matchSession(s, text)).To(Equal(expected))
		},
		Entry("matches summary", "unit tests", true),
		Entry("matches first prompt", "write tests", true),
		Entry("matches project path", "myproject", true),
		Entry("matches branch", "auth", true),
		Entry("no match", "nonexistent", false),
		Entry("case insensitive", "UNIT TESTS", true),
	)

	DescribeTable("project prefix",
		func(text string, expected bool) {
			Expect(matchSession(s, text)).To(Equal(expected))
		},
		Entry("matches basename", "project:myproject", true),
		Entry("matches full path", "project:/home/user", true),
		Entry("no match", "project:otherproject", false),
		Entry("case insensitive", "project:MYPROJECT", true),
	)

	DescribeTable("branch prefix",
		func(text string, expected bool) {
			Expect(matchSession(s, text)).To(Equal(expected))
		},
		Entry("matches", "branch:feature/auth", true),
		Entry("partial match", "branch:auth", true),
		Entry("no match", "branch:main", false),
		Entry("case insensitive", "branch:FEATURE", true),
	)

	DescribeTable("cwd prefix",
		func(text string, expected bool) {
			Expect(matchSession(s, text)).To(Equal(expected))
		},
		Entry("matches path", "cwd:/home/user", true),
		Entry("partial match", "cwd:myproject", true),
		Entry("no match", "cwd:/other/path", false),
		Entry("case insensitive", "cwd:HOME", true),
	)

	DescribeTable("pr prefix",
		func(text string, expected bool) {
			Expect(matchSession(s, text)).To(Equal(expected))
		},
		Entry("matches repo", "pr:user/myproject", true),
		Entry("matches number", "pr:42", true),
		Entry("matches repo#number", "pr:myproject#42", true),
		Entry("no match", "pr:99", false),
		Entry("case insensitive", "pr:USER/MYPROJECT", true),
	)

	DescribeTable("unknown prefix",
		func(text string, expected bool) {
			Expect(matchSession(s, text)).To(Equal(expected))
		},
		Entry("falls through to bare search", "foo:bar", false),
		Entry("no match", "Go:whatever", false),
	)

	DescribeTable("combined filters",
		func(text string, expected bool) {
			Expect(matchSession(s, text)).To(Equal(expected))
		},
		Entry("project and branch", "project:myproject branch:auth", true),
		Entry("project and branch mismatch", "project:myproject branch:main", false),
		Entry("bare and prefix", "unit project:myproject", true),
		Entry("bare and prefix mismatch", "unit project:other", false),
		Entry("multiple bare terms", "unit Go", true),
		Entry("multiple bare terms one missing", "unit Python", false),
	)

	DescribeTable("regex patterns",
		func(text string, expected bool) {
			Expect(matchSession(s, text)).To(Equal(expected))
		},
		Entry("anchor end", "project:myproject$", true),
		Entry("anchor end excludes partial", "project:myprojec$", false),
		Entry("anchor start", "project:^myproject", true),
		Entry("dot star", "project:my.*ject", true),
		Entry("bare matches field end", "tests\\?$", true),
		Entry("bare no match", "^Python", false),
		Entry("branch anchor", "branch:auth$", true),
		Entry("branch excludes", "branch:^auth$", false),
		Entry("pr number", "pr:#42$", true),
		Entry("invalid regex falls back to substring", "project:myproject", true),
		Entry("invalid regex bracket", "project:my[", false),
	)
})

var _ = Describe("matchSessions", func() {
	var sessions []claude.Session

	BeforeEach(func() {
		sessions = []claude.Session{
			{Summary: "Auth work", ProjectPath: "/home/user/auth", GitBranch: "main"},
			{Summary: "API work", ProjectPath: "/home/user/api", GitBranch: "develop"},
			{Summary: "UI work", ProjectPath: "/home/user/frontend", GitBranch: "main"},
		}
	})

	It("filters by project", func() {
		got := matchSessions(sessions, "project:auth")
		Expect(got).To(HaveLen(1))
		Expect(got[0].Summary).To(Equal("Auth work"))
	})

	It("filters by branch", func() {
		Expect(matchSessions(sessions, "branch:main")).To(HaveLen(2))
	})

	It("searches all fields with bare text", func() {
		Expect(matchSessions(sessions, "work")).To(HaveLen(3))
	})

	It("returns empty for no matches", func() {
		Expect(matchSessions(sessions, "branch:release")).To(BeEmpty())
	})

	It("narrows with combined filters", func() {
		got := matchSessions(sessions, "project:auth branch:main")
		Expect(got).To(HaveLen(1))
		Expect(got[0].Summary).To(Equal("Auth work"))
	})

	It("returns empty for combined filters with no overlap", func() {
		Expect(matchSessions(sessions, "project:api branch:main")).To(BeEmpty())
	})
})
