package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go-scraper-engine/internal/config"

	"github.com/chromedp/chromedp"
)

// FetchDynamicHTML uses a headless Chrome browser to fully render a page,
// including infinite-scroll content, and returns the final HTML string.
func FetchDynamicHTML(url string, cfg *config.Config) (string, error) {
	logger := slog.Default()

	// --- 1. Configure headless browser ---
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent(cfg.Scraper.UserAgent),
		chromedp.WindowSize(cfg.Scraper.ViewportWidth, cfg.Scraper.ViewportHeight),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	// Create browser context with the overall scroll timeout
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(logger.Info))
	defer cancel()

	// Apply scroll timeout to the entire operation
	scrollTimeout := cfg.Scraper.ScrollTimeout
	if scrollTimeout <= 0 {
		scrollTimeout = 30 * time.Second
	}
	ctx, cancel = context.WithTimeout(ctx, scrollTimeout)
	defer cancel()

	// --- 2. Navigate to URL ---
	logger.Info("Navigating to URL", slog.String("url", url))
	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		return "", fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	// Wait for initial page load
	if err := chromedp.Run(ctx, chromedp.WaitReady("body")); err != nil {
		return "", fmt.Errorf("failed waiting for body ready: %w", err)
	}

	// --- 3. Infinite Scroll Loop ---
	scrollDelay := cfg.Scraper.ScrollDelay
	if scrollDelay <= 0 {
		scrollDelay = 2 * time.Second
	}

	logger.Info("Starting infinite scroll loop",
		slog.Duration("scroll_timeout", scrollTimeout),
		slog.Duration("scroll_delay", scrollDelay),
	)

	var previousHeight int64
	scrollCount := 0

	for {
		// Get current document height
		var currentHeight int64
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`document.body.scrollHeight`, &currentHeight),
		); err != nil {
			// Context deadline exceeded means we hit ScrollTimeout — this is expected
			if ctx.Err() != nil {
				logger.Info("Scroll timeout reached, extracting current content")
				break
			}
			return "", fmt.Errorf("failed to get scroll height: %w", err)
		}

		// Check if we've reached the bottom (no new content loaded)
		if currentHeight == previousHeight {
			logger.Info("Finished scrolling — no new content detected",
				slog.Int("total_scrolls", scrollCount),
				slog.Int64("final_height", currentHeight),
			)
			break
		}

		previousHeight = currentHeight
		scrollCount++

		// Scroll to bottom
		logger.Info("Scrolling...",
			slog.Int("scroll_count", scrollCount),
			slog.Int64("current_height", currentHeight),
		)

		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil),
		); err != nil {
			if ctx.Err() != nil {
				logger.Info("Scroll timeout reached during scroll, extracting current content")
				break
			}
			return "", fmt.Errorf("failed to scroll: %w", err)
		}

		// Wait for new content to load
		time.Sleep(scrollDelay)
	}

	// --- 4. Extract full HTML ---
	// Create a fresh context for extraction in case the scroll context timed out
	extractCtx, extractCancel := chromedp.NewContext(allocCtx)
	defer extractCancel()
	extractCtx, extractCancel = context.WithTimeout(extractCtx, 15*time.Second)
	defer extractCancel()

	// Navigate again only if the original context expired
	var htmlContent string
	if ctx.Err() != nil {
		// Original context timed out, use the still-running browser's tab
		// We need to extract from the original tab, so let's try with a generous timeout
		// Re-create context from allocator (reuses browser)
		logger.Info("Re-establishing context for HTML extraction")
		if err := chromedp.Run(extractCtx,
			chromedp.Navigate(url),
			chromedp.WaitReady("body"),
		); err != nil {
			return "", fmt.Errorf("failed to re-navigate for extraction: %w", err)
		}
		if err := chromedp.Run(extractCtx,
			chromedp.OuterHTML("html", &htmlContent),
		); err != nil {
			return "", fmt.Errorf("failed to extract HTML after re-navigation: %w", err)
		}
	} else {
		if err := chromedp.Run(ctx,
			chromedp.OuterHTML("html", &htmlContent),
		); err != nil {
			return "", fmt.Errorf("failed to extract HTML: %w", err)
		}
	}

	logger.Info("HTML extraction complete",
		slog.Int("html_length", len(htmlContent)),
		slog.Int("total_scrolls", scrollCount),
	)

	return htmlContent, nil
}
