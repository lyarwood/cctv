package claude_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lyarwood/cctv/internal/claude"
)

var _ = Describe("NormalizeModel", func() {
	It("strips date suffixes", func() {
		Expect(claude.NormalizeModel("claude-opus-4-5-20251101")).To(Equal("claude-opus-4-5"))
		Expect(claude.NormalizeModel("claude-sonnet-4-5-20250929")).To(Equal("claude-sonnet-4-5"))
		Expect(claude.NormalizeModel("claude-haiku-4-5-20251001")).To(Equal("claude-haiku-4-5"))
		Expect(claude.NormalizeModel("claude-3-5-haiku-20241022")).To(Equal("claude-3-5-haiku"))
	})

	It("leaves models without date suffixes unchanged", func() {
		Expect(claude.NormalizeModel("claude-opus-4-6")).To(Equal("claude-opus-4-6"))
		Expect(claude.NormalizeModel("sonnet")).To(Equal("sonnet"))
	})

	It("trims whitespace", func() {
		Expect(claude.NormalizeModel("  claude-opus-4-6  ")).To(Equal("claude-opus-4-6"))
	})
})

var _ = Describe("LookupPricing", func() {
	It("finds pricing for known models", func() {
		p, ok := claude.LookupPricing("claude-opus-4-6")
		Expect(ok).To(BeTrue())
		Expect(p.InputPerMTok).To(Equal(5.0))
		Expect(p.CacheHitPerMTok).To(Equal(0.5))
	})

	It("finds pricing for models with date suffixes", func() {
		p, ok := claude.LookupPricing("claude-opus-4-5-20251101")
		Expect(ok).To(BeTrue())
		Expect(p.InputPerMTok).To(Equal(5.0))
	})

	It("finds pricing for sonnet alias", func() {
		p, ok := claude.LookupPricing("sonnet")
		Expect(ok).To(BeTrue())
		Expect(p.InputPerMTok).To(Equal(3.0))
	})

	It("returns false for unknown models", func() {
		_, ok := claude.LookupPricing("<synthetic>")
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("EstimateResumeCost", func() {
	It("estimates cost for opus 4.6 at 200K tokens", func() {
		cold, warm := claude.EstimateResumeCost("claude-opus-4-6", 200_000)
		Expect(cold).To(BeNumerically("~", 1.0, 0.001))
		Expect(warm).To(BeNumerically("~", 0.1, 0.001))
	})

	It("estimates cost for sonnet at 100K tokens", func() {
		cold, warm := claude.EstimateResumeCost("claude-sonnet-4-5-20250929", 100_000)
		Expect(cold).To(BeNumerically("~", 0.3, 0.001))
		Expect(warm).To(BeNumerically("~", 0.03, 0.001))
	})

	It("returns zero for unknown models", func() {
		cold, warm := claude.EstimateResumeCost("unknown-model", 100_000)
		Expect(cold).To(Equal(0.0))
		Expect(warm).To(Equal(0.0))
	})

	It("returns zero for zero tokens", func() {
		cold, warm := claude.EstimateResumeCost("claude-opus-4-6", 0)
		Expect(cold).To(Equal(0.0))
		Expect(warm).To(Equal(0.0))
	})
})
