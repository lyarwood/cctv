package claude_test

import (
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lyarwood/cctv/internal/claude"
)

var _ = Describe("Content Search", func() {
	var testdataDir string

	BeforeEach(func() {
		testdataDir = "testdata"
	})

	Describe("SearchJSONL", func() {
		It("finds matching user content", func() {
			pattern := regexp.MustCompile("(?i)meaning of life")
			result := claude.SearchJSONL(filepath.Join(testdataDir, "session-basic.jsonl"), pattern)
			Expect(result).NotTo(BeNil())
			Expect(result.Snippet).To(ContainSubstring("meaning of life"))
			Expect(result.MatchLen).To(Equal(len("meaning of life")))
		})

		It("finds matching assistant content", func() {
			pattern := regexp.MustCompile("(?i)meaning of life is 42")
			result := claude.SearchJSONL(filepath.Join(testdataDir, "session-basic.jsonl"), pattern)
			Expect(result).NotTo(BeNil())
			Expect(result.Snippet).To(ContainSubstring("meaning of life is 42"))
		})

		It("returns nil when no match", func() {
			pattern := regexp.MustCompile("(?i)nonexistent_string_xyz")
			result := claude.SearchJSONL(filepath.Join(testdataDir, "session-basic.jsonl"), pattern)
			Expect(result).To(BeNil())
		})

		It("matches with regex patterns", func() {
			pattern := regexp.MustCompile("(?i)meaning.*life")
			result := claude.SearchJSONL(filepath.Join(testdataDir, "session-basic.jsonl"), pattern)
			Expect(result).NotTo(BeNil())
			Expect(result.Snippet).To(ContainSubstring("meaning of life"))
		})

		It("handles structured content blocks", func() {
			pattern := regexp.MustCompile("(?i)structured content")
			result := claude.SearchJSONL(filepath.Join(testdataDir, "session-structured-content.jsonl"), pattern)
			Expect(result).NotTo(BeNil())
			Expect(result.Snippet).To(ContainSubstring("Structured content"))
		})

		It("returns nil for missing file", func() {
			pattern := regexp.MustCompile("(?i)test")
			result := claude.SearchJSONL("nonexistent.jsonl", pattern)
			Expect(result).To(BeNil())
		})

		It("sets correct match position in snippet", func() {
			pattern := regexp.MustCompile("(?i)elaborate")
			result := claude.SearchJSONL(filepath.Join(testdataDir, "session-basic.jsonl"), pattern)
			Expect(result).NotTo(BeNil())
			runes := []rune(result.Snippet)
			matched := string(runes[result.MatchPos : result.MatchPos+result.MatchLen])
			Expect(matched).To(Equal("elaborate"))
		})
	})

	Describe("SearchAll", func() {
		It("searches across multiple sessions", func() {
			tmpDir := GinkgoT().TempDir()
			jsonl1 := filepath.Join(tmpDir, "session1.jsonl")
			jsonl2 := filepath.Join(tmpDir, "session2.jsonl")

			Expect(writeFile(jsonl1, []byte(`{"type":"user","message":{"role":"user","content":"hello world"},"timestamp":"2026-01-01T00:00:00.000Z","sessionId":"s1","cwd":"/tmp","gitBranch":"main"}`))).To(Succeed())
			Expect(writeFile(jsonl2, []byte(`{"type":"user","message":{"role":"user","content":"goodbye world"},"timestamp":"2026-01-01T00:00:00.000Z","sessionId":"s2","cwd":"/tmp","gitBranch":"main"}`))).To(Succeed())

			sessions := []claude.Session{
				{SessionID: "s1", HasJSONL: true, FullPath: jsonl1},
				{SessionID: "s2", HasJSONL: true, FullPath: jsonl2},
			}

			d := claude.NewDiscoverer(tmpDir)
			results, err := d.SearchAll(sessions, "world", 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(2))
		})

		It("respects maxResults cap", func() {
			tmpDir := GinkgoT().TempDir()
			jsonl1 := filepath.Join(tmpDir, "session1.jsonl")
			jsonl2 := filepath.Join(tmpDir, "session2.jsonl")

			Expect(writeFile(jsonl1, []byte(`{"type":"user","message":{"role":"user","content":"hello world"},"timestamp":"2026-01-01T00:00:00.000Z","sessionId":"s1","cwd":"/tmp","gitBranch":"main"}`))).To(Succeed())
			Expect(writeFile(jsonl2, []byte(`{"type":"user","message":{"role":"user","content":"goodbye world"},"timestamp":"2026-01-01T00:00:00.000Z","sessionId":"s2","cwd":"/tmp","gitBranch":"main"}`))).To(Succeed())

			sessions := []claude.Session{
				{SessionID: "s1", HasJSONL: true, FullPath: jsonl1},
				{SessionID: "s2", HasJSONL: true, FullPath: jsonl2},
			}

			d := claude.NewDiscoverer(tmpDir)
			results, err := d.SearchAll(sessions, "world", 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))
		})

		It("skips sessions without JSONL", func() {
			sessions := []claude.Session{
				{SessionID: "s1", HasJSONL: false},
			}

			d := claude.NewDiscoverer(GinkgoT().TempDir())
			results, err := d.SearchAll(sessions, "anything", 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(BeEmpty())
		})

		It("falls back to substring for invalid regex", func() {
			tmpDir := GinkgoT().TempDir()
			jsonl := filepath.Join(tmpDir, "session.jsonl")
			Expect(writeFile(jsonl, []byte(`{"type":"user","message":{"role":"user","content":"test [bracket"},"timestamp":"2026-01-01T00:00:00.000Z","sessionId":"s1","cwd":"/tmp","gitBranch":"main"}`))).To(Succeed())

			sessions := []claude.Session{
				{SessionID: "s1", HasJSONL: true, FullPath: jsonl},
			}

			d := claude.NewDiscoverer(tmpDir)
			results, err := d.SearchAll(sessions, "[bracket", 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))
		})

		It("populates session in results", func() {
			tmpDir := GinkgoT().TempDir()
			jsonl := filepath.Join(tmpDir, "session.jsonl")
			Expect(writeFile(jsonl, []byte(`{"type":"user","message":{"role":"user","content":"findme"},"timestamp":"2026-01-01T00:00:00.000Z","sessionId":"s1","cwd":"/tmp","gitBranch":"main"}`))).To(Succeed())

			sessions := []claude.Session{
				{SessionID: "s1", HasJSONL: true, FullPath: jsonl, ProjectPath: "/some/path"},
			}

			d := claude.NewDiscoverer(tmpDir)
			results, err := d.SearchAll(sessions, "findme", 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))
			Expect(results[0].Session.SessionID).To(Equal("s1"))
			Expect(results[0].Session.ProjectPath).To(Equal("/some/path"))
		})
	})
})
