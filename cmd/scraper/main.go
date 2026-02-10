package main

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"go-scraper-engine/internal/config"
	"go-scraper-engine/internal/fetcher"
	"go-scraper-engine/internal/models"
	"go-scraper-engine/internal/storage"

	"github.com/gocolly/colly/v2"
)

// whitespaceRegex matches multiple whitespace characters (spaces, tabs, newlines).
var whitespaceRegex = regexp.MustCompile(`[\s\t\n\r]+`)

// cleanText removes extra whitespace, newlines, and tabs from text.
func cleanText(s string) string {
	cleaned := whitespaceRegex.ReplaceAllString(s, " ")
	return strings.TrimSpace(cleaned)
}

func main() {
	// Initialize structured JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Run all sites
	if err := run(logger, cfg); err != nil {
		logger.Error("scraper failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger, cfg *config.Config) error {
	logger.Info("Starting scraper engine",
		slog.String("app", cfg.App.Name),
		slog.Int("total_sites", len(cfg.Sites)),
	)

	var wg sync.WaitGroup
	errCh := make(chan error, len(cfg.Sites))

	for _, site := range cfg.Sites {
		wg.Add(1)
		go func(site config.SiteConfig) {
			defer wg.Done()
			if err := scrapeSite(logger, cfg, site); err != nil {
				errCh <- fmt.Errorf("site %s: %w", site.Name, err)
			}
		}(site)
	}

	wg.Wait()
	close(errCh)

	// Collect errors
	var errs []string
	for err := range errCh {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("scraping errors: %s", strings.Join(errs, "; "))
	}

	logger.Info("all sites scraping completed")
	return nil
}

// scrapeSite handles scraping for a single site configuration.
func scrapeSite(logger *slog.Logger, cfg *config.Config, site config.SiteConfig) error {
	siteLogger := logger.With(slog.String("site", site.Name))

	siteLogger.Info("Starting scraper for target",
		slog.String("url", site.URL),
		slog.Int("max_pages", cfg.Scraper.MaxPages),
	)

	// Dynamic output filename based on site name
	filename := fmt.Sprintf("output_%s.csv", site.Name)
	writer, err := storage.NewCSVWriter(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV writer: %w", err)
	}
	defer writer.Close()

	// Page counter for pagination limit (atomic for thread safety)
	var pageCounter int32 = 1

	// New collector per site to isolate cookies/sessions
	collector, err := fetcher.NewScraper(fetcher.ScraperOptions{
		UserAgent:   cfg.Scraper.UserAgent,
		MaxDepth:    cfg.Scraper.MaxDepth,
		Parallelism: cfg.Scraper.Parallelism,
		Delay:       cfg.Scraper.Delay,
	})
	if err != nil {
		return fmt.Errorf("failed to create scraper: %w", err)
	}

	// OnHTML callback for product extraction using site-specific selectors
	collector.OnHTML(site.Selectors.Container, func(e *colly.HTMLElement) {
		product := models.Product{
			ID:     cleanText(e.Attr("data-id")),
			Name:   cleanText(e.ChildText(site.Selectors.Title)),
			Price:  cleanText(e.ChildText(site.Selectors.Price)),
			Rating: cleanText(e.ChildText(".rating")),
			URL:    e.Request.AbsoluteURL(e.ChildAttr(site.Selectors.URL, "href")),
		}

		if err := writer.Write(product); err != nil {
			siteLogger.Error("failed to write product to CSV",
				slog.String("name", product.Name),
				slog.String("error", err.Error()),
			)
			return
		}

		siteLogger.Info("Saved",
			slog.String("name", product.Name),
			slog.String("price", product.Price),
		)
	})

	// OnHTML callback for pagination using site-specific selector
	collector.OnHTML(site.Selectors.Pagination, func(e *colly.HTMLElement) {
		nextPageURL := e.Attr("href")
		if nextPageURL == "" {
			return
		}

		currentPage := atomic.LoadInt32(&pageCounter)
		if int(currentPage) >= cfg.Scraper.MaxPages {
			siteLogger.Info("Reached max pages limit",
				slog.Int("max_pages", cfg.Scraper.MaxPages),
			)
			return
		}

		atomic.AddInt32(&pageCounter, 1)

		absoluteURL := e.Request.AbsoluteURL(nextPageURL)
		siteLogger.Info("Visiting next page",
			slog.String("url", absoluteURL),
			slog.Int("page", int(currentPage)+1),
		)

		if err := e.Request.Visit(absoluteURL); err != nil {
			siteLogger.Error("failed to visit next page",
				slog.String("url", absoluteURL),
				slog.String("error", err.Error()),
			)
		}
	})

	// Error handling callback
	collector.OnError(func(r *colly.Response, err error) {
		siteLogger.Error("request failed",
			slog.String("url", r.Request.URL.String()),
			slog.Int("status_code", r.StatusCode),
			slog.String("error", err.Error()),
		)
	})

	// Request callback for logging
	collector.OnRequest(func(r *colly.Request) {
		siteLogger.Info("visiting", slog.String("url", r.URL.String()))
	})

	// Start scraping
	if err := collector.Visit(site.URL); err != nil {
		return fmt.Errorf("failed to visit %s: %w", site.URL, err)
	}

	collector.Wait()

	siteLogger.Info("scraping completed",
		slog.String("output_file", filename),
		slog.Int("pages_scraped", int(atomic.LoadInt32(&pageCounter))),
	)
	return nil
}
