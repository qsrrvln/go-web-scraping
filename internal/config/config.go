package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the scraper application.
type Config struct {
	App     AppConfig     `mapstructure:"app"`
	Scraper ScraperConfig `mapstructure:"scraper"`
	Sites   []SiteConfig  `mapstructure:"sites"`
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name    string        `mapstructure:"name"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// ScraperConfig holds shared scraper settings.
type ScraperConfig struct {
	UserAgent      string        `mapstructure:"user_agent"`
	MaxDepth       int           `mapstructure:"max_depth"`
	Parallelism    int           `mapstructure:"parallelism"`
	Delay          time.Duration `mapstructure:"delay"`
	MaxPages       int           `mapstructure:"max_pages"`
	UseRenderer    bool          `mapstructure:"use_renderer"`
	ScrollTimeout  time.Duration `mapstructure:"scroll_timeout"`
	ScrollDelay    time.Duration `mapstructure:"scroll_delay"`
	ViewportWidth  int           `mapstructure:"viewport_width"`
	ViewportHeight int           `mapstructure:"viewport_height"`
}

// SiteConfig holds per-site scraping configuration.
type SiteConfig struct {
	Name      string          `mapstructure:"name"`
	URL       string          `mapstructure:"url"`
	Selectors SelectorsConfig `mapstructure:"selectors"`
}

// SelectorsConfig holds CSS selectors for data extraction.
type SelectorsConfig struct {
	Container  string `mapstructure:"container"`
	Title      string `mapstructure:"title"`
	Price      string `mapstructure:"price"`
	URL        string `mapstructure:"url"`
	Pagination string `mapstructure:"pagination"`
}

// LoadConfig reads configuration from the specified path.
// It looks for a file named "scraper.yaml" in the configs directory.
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("scraper")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path + "/configs")
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")

	// Set defaults for shared scraper settings
	viper.SetDefault("app.name", "go-scraper-engine")
	viper.SetDefault("app.timeout", "60s")
	viper.SetDefault("scraper.max_depth", 2)
	viper.SetDefault("scraper.parallelism", 2)
	viper.SetDefault("scraper.delay", "2s")
	viper.SetDefault("scraper.max_pages", 5)
	viper.SetDefault("scraper.use_renderer", false)
	viper.SetDefault("scraper.scroll_timeout", "30s")
	viper.SetDefault("scraper.scroll_delay", "2s")
	viper.SetDefault("scraper.viewport_width", 1920)
	viper.SetDefault("scraper.viewport_height", 1080)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if len(cfg.Sites) == 0 {
		return nil, fmt.Errorf("no sites configured in config file")
	}

	return &cfg, nil
}
