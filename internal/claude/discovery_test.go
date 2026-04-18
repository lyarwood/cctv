package claude_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lyarwood/cctv/internal/claude"
)

var _ = Describe("Discoverer", func() {
	var (
		tmpDir     string
		discoverer *claude.Discoverer
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(tmpDir, "projects", "test-project"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(tmpDir, "sessions"), 0o755)).To(Succeed())
		discoverer = claude.NewDiscoverer(tmpDir)
	})

	It("discovers sessions from index", func() {
		index := claude.IndexFile{
			Version: 1,
			Entries: []claude.IndexEntry{
				{
					SessionID:    "idx-session-1",
					FullPath:     filepath.Join(tmpDir, "projects", "test-project", "idx-session-1.jsonl"),
					FirstPrompt:  "Hello from index",
					Summary:      "Index session",
					MessageCount: 5,
					Created:      "2026-03-01T10:00:00.000Z",
					Modified:     "2026-03-01T11:00:00.000Z",
					GitBranch:    "main",
					ProjectPath:  "/home/user/project",
				},
			},
		}
		data, _ := json.Marshal(index)
		Expect(writeFile(filepath.Join(tmpDir, "projects", "test-project", "sessions-index.json"), data)).To(Succeed())

		sessions, err := discoverer.DiscoverAll()
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(HaveLen(1))
		Expect(sessions[0].SessionID).To(Equal("idx-session-1"))
		Expect(sessions[0].Summary).To(Equal("Index session"))
	})

	It("discovers sessions from JSONL files not in index", func() {
		jsonlContent := `{"type":"permission-mode","permissionMode":"default","sessionId":"jsonl-session-1"}
{"type":"user","message":{"role":"user","content":"Hello from JSONL"},"timestamp":"2026-03-02T10:00:00.000Z","cwd":"/home/user/other","sessionId":"jsonl-session-1","version":"2.1.96","gitBranch":"develop","uuid":"u1","parentUuid":null,"isSidechain":false,"userType":"external","entrypoint":"cli"}
`
		Expect(writeFile(filepath.Join(tmpDir, "projects", "test-project", "jsonl-session-1.jsonl"), []byte(jsonlContent))).To(Succeed())

		sessions, err := discoverer.DiscoverAll()
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(HaveLen(1))
		Expect(sessions[0].SessionID).To(Equal("jsonl-session-1"))
		Expect(sessions[0].FirstPrompt).To(Equal("Hello from JSONL"))
		Expect(sessions[0].Source).To(Equal(claude.SourceJSONL))
	})

	It("marks sessions present in both index and JSONL as SourceBoth", func() {
		jsonlContent := `{"type":"permission-mode","permissionMode":"default","sessionId":"both-session"}
{"type":"user","message":{"role":"user","content":"I am in both"},"timestamp":"2026-03-03T10:00:00.000Z","cwd":"/home/user/both","sessionId":"both-session","version":"2.1.96","gitBranch":"main","uuid":"u1","parentUuid":null,"isSidechain":false,"userType":"external","entrypoint":"cli"}
`
		jsonlPath := filepath.Join(tmpDir, "projects", "test-project", "both-session.jsonl")
		Expect(writeFile(jsonlPath, []byte(jsonlContent))).To(Succeed())

		index := claude.IndexFile{
			Version: 1,
			Entries: []claude.IndexEntry{
				{
					SessionID:    "both-session",
					FullPath:     jsonlPath,
					FirstPrompt:  "I am in both",
					Summary:      "Both session",
					MessageCount: 2,
					Created:      "2026-03-03T10:00:00.000Z",
					Modified:     "2026-03-03T10:30:00.000Z",
					GitBranch:    "main",
					ProjectPath:  "/home/user/both",
				},
			},
		}
		data, _ := json.Marshal(index)
		Expect(writeFile(filepath.Join(tmpDir, "projects", "test-project", "sessions-index.json"), data)).To(Succeed())

		sessions, err := discoverer.DiscoverAll()
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(HaveLen(1))
		Expect(sessions[0].Source).To(Equal(claude.SourceBoth))
		Expect(sessions[0].HasJSONL).To(BeTrue())
	})

	It("marks running sessions", func() {
		jsonlContent := `{"type":"permission-mode","permissionMode":"default","sessionId":"running-session"}
{"type":"user","message":{"role":"user","content":"I am running"},"timestamp":"2026-03-04T10:00:00.000Z","cwd":"/home/user/run","sessionId":"running-session","version":"2.1.96","gitBranch":"main","uuid":"u1","parentUuid":null,"isSidechain":false,"userType":"external","entrypoint":"cli"}
`
		Expect(writeFile(filepath.Join(tmpDir, "projects", "test-project", "running-session.jsonl"), []byte(jsonlContent))).To(Succeed())

		pid := os.Getpid()
		rs := claude.RunningSession{
			PID:       pid,
			SessionID: "running-session",
			CWD:       "/home/user/run",
		}
		data, _ := json.Marshal(rs)
		Expect(writeFile(filepath.Join(tmpDir, "sessions", fmt.Sprintf("%d.json", pid)), data)).To(Succeed())

		sessions, err := discoverer.DiscoverAll()
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(HaveLen(1))
		Expect(sessions[0].IsRunning).To(BeTrue())
		Expect(sessions[0].RunningPID).To(Equal(pid))
	})

	It("sorts sessions by modified descending", func() {
		index := claude.IndexFile{
			Version: 1,
			Entries: []claude.IndexEntry{
				{
					SessionID:   "older",
					FirstPrompt: "old",
					Created:     "2026-01-01T10:00:00.000Z",
					Modified:    "2026-01-01T10:00:00.000Z",
					ProjectPath: "/old",
				},
				{
					SessionID:   "newer",
					FirstPrompt: "new",
					Created:     "2026-03-01T10:00:00.000Z",
					Modified:    "2026-03-01T10:00:00.000Z",
					ProjectPath: "/new",
				},
			},
		}
		data, _ := json.Marshal(index)
		Expect(writeFile(filepath.Join(tmpDir, "projects", "test-project", "sessions-index.json"), data)).To(Succeed())

		sessions, err := discoverer.DiscoverAll()
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(HaveLen(2))
		Expect(sessions[0].SessionID).To(Equal("newer"))
		Expect(sessions[1].SessionID).To(Equal("older"))
	})

	It("handles empty project directory", func() {
		sessions, err := discoverer.DiscoverAll()
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(BeEmpty())
	})
})
