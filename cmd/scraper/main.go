package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go-scraper-engine/internal/fetcher"
	"go-scraper-engine/internal/models"

	"github.com/gocolly/colly/v2"
)

func main() {
	// Initialize structured JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Run the scraper
	if err := run(logger); err != nil {
		logger.Error("scraper failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// Initialize the collector via dependency injection
	collector, err := fetcher.NewScraper()
	if err != nil {
		return fmt.Errorf("failed to create scraper: %w", err)
	}

	// Placeholder target URL (replace with actual e-commerce site)
	targetURL := "https://example.com/products"

	// Define OnHTML callback for product extraction
	// Using placeholder selectors for generic e-commerce structure
	collector.OnHTML(".product-card", func(e *colly.HTMLElement) {
		product := models.Product{
			ID:     e.Attr("data-id"),
			Name:   strings.TrimSpace(e.ChildText(".title")),
			Price:  strings.TrimSpace(e.ChildText(".price")),
			Rating: strings.TrimSpace(e.ChildText(".rating")),
			URL:    e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
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

	// Start scraping
	if err := collector.Visit(targetURL); err != nil {
		return fmt.Errorf("failed to visit %s: %w", targetURL, err)
	}

	// Wait for async collector to finish
	collector.Wait()

	logger.Info("scraping completed")
	return nil
}
