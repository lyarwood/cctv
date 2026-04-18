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

var _ = Describe("LoadRunning", func() {
	It("returns empty map when sessions dir does not exist", func() {
		tmpDir := GinkgoT().TempDir()
		result, err := claude.LoadRunning(tmpDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
	})

	It("detects a running session with a live PID", func() {
		tmpDir := GinkgoT().TempDir()
		sessionsDir := filepath.Join(tmpDir, "sessions")
		Expect(os.MkdirAll(sessionsDir, 0o755)).To(Succeed())

		// Use our own PID so the alive check passes
		pid := os.Getpid()
		rs := claude.RunningSession{
			PID:        pid,
			SessionID:  "live-session-id",
			CWD:        "/tmp",
			StartedAt:  1776499051892,
			Kind:       "interactive",
			Entrypoint: "cli",
		}
		data, err := json.Marshal(rs)
		Expect(err).NotTo(HaveOccurred())
		Expect(writeFile(filepath.Join(sessionsDir, fmt.Sprintf("%d.json", pid)), data)).To(Succeed())

		result, err := claude.LoadRunning(tmpDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveKey("live-session-id"))
		Expect(result["live-session-id"].PID).To(Equal(pid))
	})

	It("skips sessions with dead PIDs", func() {
		tmpDir := GinkgoT().TempDir()
		sessionsDir := filepath.Join(tmpDir, "sessions")
		Expect(os.MkdirAll(sessionsDir, 0o755)).To(Succeed())

		rs := claude.RunningSession{
			PID:       999999999,
			SessionID: "dead-session-id",
		}
		data, err := json.Marshal(rs)
		Expect(err).NotTo(HaveOccurred())
		Expect(writeFile(filepath.Join(sessionsDir, "999999999.json"), data)).To(Succeed())

		result, err := claude.LoadRunning(tmpDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(HaveKey("dead-session-id"))
	})
})
