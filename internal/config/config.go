package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the scraper application.
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Scraper   ScraperConfig   `mapstructure:"scraper"`
	Selectors SelectorsConfig `mapstructure:"selectors"`
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name    string        `mapstructure:"name"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// ScraperConfig holds scraper-specific configuration.
type ScraperConfig struct {
	StartURL    string        `mapstructure:"start_url"`
	UserAgent   string        `mapstructure:"user_agent"`
	MaxDepth    int           `mapstructure:"max_depth"`
	Parallelism int           `mapstructure:"parallelism"`
	Delay       time.Duration `mapstructure:"delay"`
}

// SelectorsConfig holds CSS selectors for data extraction.
type SelectorsConfig struct {
	Container string `mapstructure:"container"`
	Title     string `mapstructure:"title"`
	Price     string `mapstructure:"price"`
	URL       string `mapstructure:"url"`
}

// LoadConfig reads configuration from the specified path.
// It looks for a file named "scraper.yaml" in the configs directory.
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("scraper")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path + "/configs")
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("app.name", "go-scraper-engine")
	viper.SetDefault("app.timeout", "60s")
	viper.SetDefault("scraper.max_depth", 2)
	viper.SetDefault("scraper.parallelism", 2)
	viper.SetDefault("scraper.delay", "2s")
	viper.SetDefault("selectors.container", ".product-card")
	viper.SetDefault("selectors.title", ".title")
	viper.SetDefault("selectors.price", ".price")
	viper.SetDefault("selectors.url", "a")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
