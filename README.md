# 🕷️ Go Scraper Engine

A high-performance, configurable web scraping engine built with Go. Supports both **static HTML scraping** (via Colly) and **dynamic JavaScript-rendered pages** (via Chromedp headless browser), with automatic infinite scroll handling.

---

## ✨ Features

- **Dual Scraping Modes** — Seamlessly switch between static (Colly) and dynamic (Chromedp) scraping via a single config flag.
- **Infinite Scroll Support** — Automatically scrolls dynamic pages (e.g., Nike) and waits for new content to load.
- **YAML-Driven Configuration** — Define target sites, CSS selectors, and scraper behavior entirely from `scraper.yaml`.
- **Multi-Site Parallel Scraping** — Scrape multiple sites concurrently using goroutines.
- **Pagination Support** — Follows pagination links for static sites with configurable page limits.
- **CSV Export** — Automatically saves scraped products to per-site CSV files.
- **User-Agent Rotation** — Randomized or fixed User-Agent for anti-bot evasion.
- **Rate Limiting** — Configurable delay and parallelism to be respectful to target servers.
- **Structured Logging** — JSON-formatted logs via Go's `slog` for easy debugging and monitoring.

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────┐
│                   main.go (cmd)                  │
│         Orchestrator & Site Dispatcher            │
├──────────────┬───────────────────────────────────┤
│              │                                    │
│   ┌──────────▼──────────┐  ┌─────────────────┐   │
│   │  fetcher/colly.go   │  │ fetcher/browser  │   │
│   │   (Static Pages)    │  │  (Dynamic/SPA)   │   │
│   │   Colly Collector   │  │   Chromedp +     │   │
│   │   + Pagination      │  │ Infinite Scroll  │   │
│   └──────────┬──────────┘  └────────┬─────────┘   │
│              │                      │              │
│              └──────────┬───────────┘              │
│                         ▼                          │
│              ┌─────────────────────┐               │
│              │   models/product    │               │
│              │   Data Extraction   │               │
│              └──────────┬──────────┘               │
│                         ▼                          │
│              ┌─────────────────────┐               │
│              │   storage/csv.go    │               │
│              │    CSV Writer       │               │
│              └─────────────────────┘               │
└──────────────────────────────────────────────────┘
```

---

## 📦 Tech Stack

| Category             | Technology                                        | Purpose                                           |
| -------------------- | ------------------------------------------------- | ------------------------------------------------- |
| **Language**         | Go 1.23+                                          | Core runtime                                      |
| **Static Scraping**  | [Colly v2](https://github.com/gocolly/colly)      | HTTP requests, collector lifecycle, rate limiting |
| **Dynamic Scraping** | [Chromedp](https://github.com/chromedp/chromedp)  | Headless Chrome for JS-rendered / SPA pages       |
| **DOM Parsing**      | [GoQuery](https://github.com/PuerkitoBio/goquery) | CSS selector-based HTML traversal                 |
| **Configuration**    | [Viper](https://github.com/spf13/viper)           | YAML config management                            |
| **Logging**          | `log/slog` (stdlib)                               | Structured JSON logging                           |
| **Concurrency**      | `sync`, `channels` (stdlib)                       | Parallel site scraping                            |
| **Storage**          | `encoding/csv` (stdlib)                           | CSV file output                                   |

---

## 📁 Project Structure

```
go-scraper-engine/
├── cmd/
│   └── scraper/
│       └── main.go              # Entry point & orchestrator
├── configs/
│   └── scraper.yaml             # Scraping configuration (sites, selectors, flags)
├── docs/
│   └── tech-stack.md            # Technical documentation
├── internal/
│   ├── config/
│   │   └── config.go            # Config structs & Viper loader
│   ├── fetcher/
│   │   ├── browser.go           # Chromedp dynamic renderer + infinite scroll
│   │   └── colly.go             # Colly static scraper + rate limiting
│   ├── models/
│   │   └── product.go           # Product data model
│   └── storage/
│       └── csv.go               # CSV file writer
├── go.mod
├── go.sum
└── README.md
```

---

## 🚀 Getting Started

### Prerequisites

- **Go** 1.23 or later — [Download](https://go.dev/dl/)
- **Google Chrome** (required only for dynamic/renderer mode) — Chromedp will use the installed Chrome binary.

### Installation

```bash
# Clone the repository
git clone https://github.com/your-username/go-scraper-engine.git
cd go-scraper-engine

# Install dependencies
go mod download
```

---

## ⚙️ Configuration

All scraper behavior is controlled via `configs/scraper.yaml`:

```yaml
app:
  name: "go-scraper-engine"
  timeout: 60s

scraper:
  user_agent: "Mozilla/5.0 ..." # Custom User-Agent (leave empty for random rotation)
  max_depth: 2 # Max link-following depth (Colly mode)
  parallelism: 2 # Concurrent requests per domain
  delay: 2s # Delay between requests
  max_pages: 5 # Pagination limit (Colly mode)
  use_renderer: false # false = Colly (static), true = Chromedp (dynamic)
  scroll_timeout: 30s # Max time for infinite scrolling (Chromedp mode)
  scroll_delay: 2s # Wait time between scrolls
  viewport_width: 1920 # Browser viewport width
  viewport_height: 1080 # Browser viewport height

sites:
  - name: "Nike Etalase (Listing)"
    url: "https://www.nike.com/id/w/mens-shoes-nik1zy7ok"
    selectors:
      container: "[data-testid='product-card__body']"
      title: ".product-card__title"
      price: "[data-testid='product-price']"
      url: "a.product-card__link-overlay"
      pagination: "" # Empty = no pagination (use infinite scroll)
```

### Key Config Options

| Option           | Description                                                         | Default |
| ---------------- | ------------------------------------------------------------------- | ------- |
| `use_renderer`   | `false` → Colly (static HTML), `true` → Chromedp (headless browser) | `false` |
| `scroll_timeout` | Maximum duration for the infinite scroll loop                       | `30s`   |
| `scroll_delay`   | Pause between each scroll action (wait for content)                 | `2s`    |
| `max_pages`      | Maximum pages to follow via pagination (Colly mode only)            | `5`     |
| `parallelism`    | Number of concurrent requests per domain                            | `2`     |
| `delay`          | Delay between consecutive requests                                  | `2s`    |

---

## ▶️ Usage

### Run the Scraper

```bash
go run ./cmd/scraper
```

### Build & Run

```bash
go build -o scraper ./cmd/scraper
./scraper
```

### Output

Scraped data is saved as CSV files in the project root, named per site:

```
output_Nike Etalase (Listing).csv
output_PracticeSite1.csv
```

**CSV format:**

| ID  | Name            | Price        | URL                           |
| --- | --------------- | ------------ | ----------------------------- |
|     | Nike Air Max 90 | Rp 1.999.000 | https://www.nike.com/id/t/... |

---

## 🔄 How It Works

### Static Mode (Colly)

1. Sends HTTP requests to the target URL.
2. Parses the HTML response using CSS selectors.
3. Extracts product data (name, price, URL).
4. Follows pagination links up to `max_pages`.
5. Writes results to CSV.

### Dynamic Mode (Chromedp)

1. Launches a headless Chrome browser.
2. Navigates to the target URL and waits for full DOM ready.
3. Performs infinite scrolling:
   - Scrolls to bottom → waits `scroll_delay` → checks if new content loaded.
   - Repeats until no new content **or** `scroll_timeout` is reached.
4. Extracts the fully rendered HTML.
5. Parses with GoQuery using the same CSS selectors.
6. Writes results to CSV.

---

## 🛠️ Adding a New Site

1. Open `configs/scraper.yaml`.
2. Add a new entry under `sites`:

```yaml
sites:
  - name: "My Store"
    url: "https://example.com/products"
    selectors:
      container: ".product-card" # Selector for each product container
      title: ".product-title" # Selector for product name (within container)
      price: ".product-price" # Selector for product price
      url: "a.product-link" # Selector for product link
      pagination: "a.next-page" # Pagination link selector (leave "" for infinite scroll)
```

3. Set `use_renderer: true` if the site requires JavaScript rendering.
4. Run the scraper — output will be saved as `output_My Store.csv`.

---

## 📝 License

This project is for educational and personal use.
