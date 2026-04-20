package claude_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lyarwood/cctv/internal/claude"
)

var _ = Describe("JSONL Parsing", func() {
	var testdataDir string

	BeforeEach(func() {
		testdataDir = "testdata"
	})

	Describe("ParseJSONLMetadata", func() {
		It("extracts metadata from the first user message", func() {
			session, err := claude.ParseJSONLMetadata(filepath.Join(testdataDir, "session-basic.jsonl"))
			Expect(err).NotTo(HaveOccurred())

			Expect(session.SessionID).To(Equal("dddd4444-dddd-dddd-dddd-dddddddddddd"))
			Expect(session.FirstPrompt).To(Equal("What is the meaning of life?"))
			Expect(session.ProjectPath).To(Equal("/home/user/philosophy"))
			Expect(session.GitBranch).To(Equal("main"))
			Expect(session.HasJSONL).To(BeTrue())
			Expect(session.Source).To(Equal(claude.SourceJSONL))
			Expect(session.Created).NotTo(BeZero())
		})

		It("aggregates total tokens from assistant messages", func() {
			session, err := claude.ParseJSONLMetadata(filepath.Join(testdataDir, "session-basic.jsonl"))
			Expect(err).NotTo(HaveOccurred())
			// msg1: 100 input + 50 output = 150; msg2: 150 input + 75 output = 225; total = 375
			Expect(session.TotalTokens).To(Equal(int64(375)))
		})

		It("handles structured content blocks", func() {
			session, err := claude.ParseJSONLMetadata(filepath.Join(testdataDir, "session-structured-content.jsonl"))
			Expect(err).NotTo(HaveOccurred())

			Expect(session.SessionID).To(Equal("eeee5555-eeee-eeee-eeee-eeeeeeeeeeee"))
			Expect(session.FirstPrompt).To(Equal("Structured content message"))
		})

		It("returns error for missing file", func() {
			_, err := claude.ParseJSONLMetadata("nonexistent.jsonl")
			Expect(err).To(HaveOccurred())
		})

		It("extracts deduplicated PR links", func() {
			session, err := claude.ParseJSONLMetadata(filepath.Join(testdataDir, "session-basic.jsonl"))
			Expect(err).NotTo(HaveOccurred())

			Expect(session.PRLinks).To(HaveLen(2))
			Expect(session.PRLinks[0].Repository).To(Equal("user/philosophy"))
			Expect(session.PRLinks[0].Number).To(Equal(42))
			Expect(session.PRLinks[0].URL).To(Equal("https://github.com/user/philosophy/pull/42"))
			Expect(session.PRLinks[1].Number).To(Equal(99))
		})

		It("returns empty PR links when none present", func() {
			session, err := claude.ParseJSONLMetadata(filepath.Join(testdataDir, "session-structured-content.jsonl"))
			Expect(err).NotTo(HaveOccurred())

			Expect(session.PRLinks).To(BeEmpty())
		})

		It("returns error for JSONL with no user messages", func() {
			tmpDir := GinkgoT().TempDir()
			path := filepath.Join(tmpDir, "empty.jsonl")
			Expect(writeFile(path, []byte(`{"type":"permission-mode","permissionMode":"default","sessionId":"test"}`))).To(Succeed())

			_, err := claude.ParseJSONLMetadata(path)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no user message found"))
		})
	})

	Describe("ParseJSONLDetail", func() {
		It("extracts all models used", func() {
			detail, err := claude.ParseJSONLDetail(filepath.Join(testdataDir, "session-basic.jsonl"))
			Expect(err).NotTo(HaveOccurred())

			Expect(detail.Models).To(ConsistOf("claude-sonnet-4-6", "claude-opus-4-6"))
		})

		It("aggregates token usage", func() {
			detail, err := claude.ParseJSONLDetail(filepath.Join(testdataDir, "session-basic.jsonl"))
			Expect(err).NotTo(HaveOccurred())

			Expect(detail.TotalUsage.InputTokens).To(Equal(int64(250)))
			Expect(detail.TotalUsage.OutputTokens).To(Equal(int64(125)))
			Expect(detail.TotalUsage.CacheCreationInputTokens).To(Equal(int64(200)))
			Expect(detail.TotalUsage.CacheReadInputTokens).To(Equal(int64(300)))
		})

		It("collects all user prompts", func() {
			detail, err := claude.ParseJSONLDetail(filepath.Join(testdataDir, "session-basic.jsonl"))
			Expect(err).NotTo(HaveOccurred())

			Expect(detail.Prompts).To(HaveLen(2))
			Expect(detail.Prompts[0].Content).To(Equal("What is the meaning of life?"))
			Expect(detail.Prompts[1].Content).To(Equal("Can you elaborate?"))
		})

		It("extracts the last prompt", func() {
			detail, err := claude.ParseJSONLDetail(filepath.Join(testdataDir, "session-basic.jsonl"))
			Expect(err).NotTo(HaveOccurred())

			Expect(detail.LastPrompt).To(Equal("Can you elaborate?"))
		})

		It("extracts the Claude Code version", func() {
			detail, err := claude.ParseJSONLDetail(filepath.Join(testdataDir, "session-basic.jsonl"))
			Expect(err).NotTo(HaveOccurred())

			Expect(detail.Version).To(Equal("2.1.96"))
		})

		It("extracts deduplicated PR links", func() {
			detail, err := claude.ParseJSONLDetail(filepath.Join(testdataDir, "session-basic.jsonl"))
			Expect(err).NotTo(HaveOccurred())

			Expect(detail.PRLinks).To(HaveLen(2))
			Expect(detail.PRLinks[0].Repository).To(Equal("user/philosophy"))
			Expect(detail.PRLinks[0].Number).To(Equal(42))
			Expect(detail.PRLinks[1].Number).To(Equal(99))
		})
	})
})
