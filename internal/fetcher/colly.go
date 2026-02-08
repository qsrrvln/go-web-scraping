package fetcher

import (
	"math/rand"
	"time"

	"github.com/gocolly/colly/v2"
)

// userAgents contains a list of common browser User-Agent strings for rotation.
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

// ScraperOptions holds configuration options for the scraper.
type ScraperOptions struct {
	UserAgent   string
	MaxDepth    int
	Parallelism int
	Delay       time.Duration
}

// getEffectiveUserAgent returns the configured user agent or a random one from the pool.
func getEffectiveUserAgent(configured string) string {
	if configured != "" {
		return configured
	}
	return userAgents[rand.Intn(len(userAgents))]
}

// NewScraper creates and returns a configured colly.Collector with:
// - Async mode enabled
// - Configurable or random User-Agent
// - Configurable rate limiting
func NewScraper(opts ScraperOptions) (*colly.Collector, error) {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent(getEffectiveUserAgent(opts.UserAgent)),
		colly.MaxDepth(opts.MaxDepth),
	)

	// Use defaults if not specified
	parallelism := opts.Parallelism
	if parallelism <= 0 {
		parallelism = 2
	}

	delay := opts.Delay
	if delay <= 0 {
		delay = 2 * time.Second
	}

	// Configure rate limiting rule
	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: parallelism,
		Delay:       delay,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
