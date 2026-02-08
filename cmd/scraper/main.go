package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go-scraper-engine/internal/config"
	"go-scraper-engine/internal/fetcher"
	"go-scraper-engine/internal/models"

	"github.com/gocolly/colly/v2"
)

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
	)

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
		product := models.Product{
			ID:     e.Attr("data-id"),
			Name:   strings.TrimSpace(e.ChildText(cfg.Selectors.Title)),
			Price:  strings.TrimSpace(e.ChildText(cfg.Selectors.Price)),
			Rating: strings.TrimSpace(e.ChildText(".rating")),
			URL:    e.Request.AbsoluteURL(e.ChildAttr(cfg.Selectors.URL, "href")),
		}

		logger.Info("product scraped",
			slog.String("id", product.ID),
			slog.String("name", product.Name),
			slog.String("price", product.Price),
			slog.String("rating", product.Rating),
			slog.String("url", product.URL),
		)
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

	logger.Info("scraping completed")
	return nil
}
