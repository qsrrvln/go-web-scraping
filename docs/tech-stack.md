## TECH STACK & DEPENDENCIES

### Core Language

- **Language**: Go (Golang)
- **Version**: Latest Stable (1.23+)

### Scraping & Parsing Engines

- **Static Scraping (Default)**: `github.com/gocolly/colly/v2`
  - _Usage_: Primary engine for handling HTTP requests, collector lifecycle, and rate limiting.
- **Dynamic Scraping (JS Rendering)**: `github.com/chromedp/chromedp`
  - _Usage_: Fallback engine strictly for SPA (Single Page Applications) or sites requiring heavy DOM manipulation via JavaScript.
- **DOM Parsing**: `github.com/PuerkitoBio/goquery`
  - _Usage_: Traversing HTML nodes using CSS selectors (jQuery-like syntax).

### Utilities & Infrastructure

- **Configuration Management**: `github.com/spf13/viper`
  - _Usage_: Externalizing selectors, target URLs, and runtime flags (YAML/JSON/Env).
- **Structured Logging**: `log/slog` (Stdlib)
  - _Usage_: JSON-formatted logging for debugging and audit trails.
- **Concurrency Control**: Standard `sync.WaitGroup` and `channels`.
  - _Usage_: Managing goroutines for parallel scraping.

### Data Persistence

- **CSV Encoding**: `encoding/csv` (Stdlib)
- **JSON Encoding**: `encoding/json` (Stdlib)
