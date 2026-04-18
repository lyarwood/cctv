package claude_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lyarwood/cctv/internal/claude"
)

var _ = Describe("ParseIndex", func() {
	var testdataDir string

	BeforeEach(func() {
		testdataDir = "testdata"
	})

	It("parses a valid sessions-index.json", func() {
		sessions, err := claude.ParseIndex(filepath.Join(testdataDir, "sessions-index.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(HaveLen(2))

		Expect(sessions[0].SessionID).To(Equal("aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa"))
		Expect(sessions[0].Summary).To(Equal("Writing unit tests in Go"))
		Expect(sessions[0].FirstPrompt).To(Equal("How do I write tests?"))
		Expect(sessions[0].ProjectPath).To(Equal("/home/user/myproject"))
		Expect(sessions[0].GitBranch).To(Equal("main"))
		Expect(sessions[0].MessageCount).To(Equal(14))
		Expect(sessions[0].Source).To(Equal(claude.SourceIndex))

		Expect(sessions[1].SessionID).To(Equal("bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb"))
		Expect(sessions[1].Summary).To(Equal("Debugging auth middleware"))
		Expect(sessions[1].GitBranch).To(Equal("fix/auth"))
	})

	It("skips sidechain entries", func() {
		sessions, err := claude.ParseIndex(filepath.Join(testdataDir, "sessions-index.json"))
		Expect(err).NotTo(HaveOccurred())
		for _, s := range sessions {
			Expect(s.SessionID).NotTo(Equal("cccc3333-cccc-cccc-cccc-cccccccccccc"))
		}
	})

	It("returns empty slice for empty entries", func() {
		sessions, err := claude.ParseIndex(filepath.Join(testdataDir, "sessions-index-empty.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(BeEmpty())
	})

	It("returns error for missing file", func() {
		_, err := claude.ParseIndex(filepath.Join(testdataDir, "nonexistent.json"))
		Expect(err).To(HaveOccurred())
	})

	It("returns error for malformed JSON", func() {
		// Create a temp file with bad JSON
		tmpDir := GinkgoT().TempDir()
		badPath := filepath.Join(tmpDir, "bad.json")
		Expect(writeFile(badPath, []byte("{invalid json"))).To(Succeed())

		_, err := claude.ParseIndex(badPath)
		Expect(err).To(HaveOccurred())
	})
})
