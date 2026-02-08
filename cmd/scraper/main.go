package main

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
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
	// Replace multiple whitespace with single space
	cleaned := whitespaceRegex.ReplaceAllString(s, " ")
	// Trim leading/trailing whitespace
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

	// Run the scraper
	if err := run(logger, cfg); err != nil {
		logger.Error("scraper failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger, cfg *config.Config) error {
	// Log startup message with target URL
	logger.Info("Starting scraper for target",
		slog.String("url", cfg.Scraper.StartURL),
		slog.String("app", cfg.App.Name),
		slog.Int("max_pages", cfg.Scraper.MaxPages),
	)

	// Initialize CSV writer for persistence
	writer, err := storage.NewCSVWriter("products.csv")
	if err != nil {
		return fmt.Errorf("failed to create CSV writer: %w", err)
	}
	defer writer.Close()

	// Page counter for pagination limit (atomic for thread safety)
	var pageCounter int32 = 1

	// Initialize the collector via dependency injection with config options
	collector, err := fetcher.NewScraper(fetcher.ScraperOptions{
		UserAgent:   cfg.Scraper.UserAgent,
		MaxDepth:    cfg.Scraper.MaxDepth,
		Parallelism: cfg.Scraper.Parallelism,
		Delay:       cfg.Scraper.Delay,
	})
	if err != nil {
		return fmt.Errorf("failed to create scraper: %w", err)
	}

	// Define OnHTML callback for product extraction using configured selectors
	collector.OnHTML(cfg.Selectors.Container, func(e *colly.HTMLElement) {
		// Extract and clean text fields
		product := models.Product{
			ID:     cleanText(e.Attr("data-id")),
			Name:   cleanText(e.ChildText(cfg.Selectors.Title)),
			Price:  cleanText(e.ChildText(cfg.Selectors.Price)),
			Rating: cleanText(e.ChildText(".rating")),
			URL:    e.Request.AbsoluteURL(e.ChildAttr(cfg.Selectors.URL, "href")),
		}

		// Write to CSV file
		if err := writer.Write(product); err != nil {
			logger.Error("failed to write product to CSV",
				slog.String("name", product.Name),
				slog.String("error", err.Error()),
			)
			return
		}

		// Log success indicator
		logger.Info("Saved",
			slog.String("name", product.Name),
			slog.String("price", product.Price),
		)
	})

	// Define OnHTML callback for pagination
	collector.OnHTML(cfg.Selectors.Pagination, func(e *colly.HTMLElement) {
		nextPageURL := e.Attr("href")
		if nextPageURL == "" {
			return
		}

		currentPage := atomic.LoadInt32(&pageCounter)
		if int(currentPage) >= cfg.Scraper.MaxPages {
			logger.Info("Reached max pages limit",
				slog.Int("max_pages", cfg.Scraper.MaxPages),
			)
			return
		}

		// Increment page counter
		atomic.AddInt32(&pageCounter, 1)

		absoluteURL := e.Request.AbsoluteURL(nextPageURL)
		logger.Info("Visiting next page",
			slog.String("url", absoluteURL),
			slog.Int("page", int(currentPage)+1),
		)

		if err := e.Request.Visit(absoluteURL); err != nil {
			logger.Error("failed to visit next page",
				slog.String("url", absoluteURL),
				slog.String("error", err.Error()),
			)
		}
	})

	// Error handling callback
	collector.OnError(func(r *colly.Response, err error) {
		logger.Error("request failed",
			slog.String("url", r.Request.URL.String()),
			slog.Int("status_code", r.StatusCode),
			slog.String("error", err.Error()),
		)
	})

	// Request callback for logging
	collector.OnRequest(func(r *colly.Request) {
		logger.Info("visiting", slog.String("url", r.URL.String()))
	})

	// Start scraping using configured start URL
	if err := collector.Visit(cfg.Scraper.StartURL); err != nil {
		return fmt.Errorf("failed to visit %s: %w", cfg.Scraper.StartURL, err)
	}

	// Wait for async collector to finish
	collector.Wait()

	logger.Info("scraping completed",
		slog.Int("pages_scraped", int(atomic.LoadInt32(&pageCounter))),
	)
	return nil
}
