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

// randomUserAgent returns a random User-Agent string from the pool.
func randomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// NewScraper creates and returns a configured colly.Collector with:
// - Async mode enabled
// - Random User-Agent rotation
// - Rate limiting (Domain Glob: "*", Parallelism: 2, Delay: 2s)
func NewScraper() (*colly.Collector, error) {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent(randomUserAgent()),
	)

	// Configure rate limiting rule
	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       2 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
