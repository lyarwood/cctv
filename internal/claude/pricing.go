package claude

import (
	"regexp"
	"strings"
)

type ModelPricing struct {
	InputPerMTok    float64
	CacheHitPerMTok float64
}

var modelPricing = map[string]ModelPricing{
	// Opus 4.5+ ($5 input, $0.50 cache hit)
	"claude-opus-4-7":   {5.00, 0.50},
	"claude-opus-4-6":   {5.00, 0.50},
	"claude-opus-4-5":   {5.00, 0.50},
	// Opus 4.0-4.1 ($15 input, $1.50 cache hit)
	"claude-opus-4-1":   {15.00, 1.50},
	"claude-opus-4":     {15.00, 1.50},
	"claude-opus-3":     {15.00, 1.50},
	// Sonnet ($3 input, $0.30 cache hit)
	"claude-sonnet-4-6": {3.00, 0.30},
	"claude-sonnet-4-5": {3.00, 0.30},
	"claude-sonnet-4":   {3.00, 0.30},
	"claude-sonnet-3-7": {3.00, 0.30},
	"sonnet":            {3.00, 0.30},
	// Haiku 4.5 ($1 input, $0.10 cache hit)
	"claude-haiku-4-5":  {1.00, 0.10},
	// Haiku 3.5 ($0.80 input, $0.08 cache hit)
	"claude-3-5-haiku":  {0.80, 0.08},
	"claude-haiku-3-5":  {0.80, 0.08},
	// Haiku 3 ($0.25 input, $0.03 cache hit)
	"claude-haiku-3":    {0.25, 0.03},
	"claude-3-haiku":    {0.25, 0.03},
}

var dateStripRe = regexp.MustCompile(`-\d{8}$`)

func NormalizeModel(model string) string {
	return dateStripRe.ReplaceAllString(strings.TrimSpace(model), "")
}

func LookupPricing(model string) (ModelPricing, bool) {
	normalized := NormalizeModel(model)
	p, ok := modelPricing[normalized]
	return p, ok
}

func EstimateResumeCost(model string, inputTokens int64) (coldCost, warmCost float64) {
	p, ok := LookupPricing(model)
	if !ok || inputTokens <= 0 {
		return 0, 0
	}
	mtok := float64(inputTokens) / 1_000_000
	return mtok * p.InputPerMTok, mtok * p.CacheHitPerMTok
}
