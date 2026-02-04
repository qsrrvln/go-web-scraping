package models

// Product represents a scraped e-commerce product item.
type Product struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Price  string `json:"price"`
	Rating string `json:"rating"`
	URL    string `json:"url"`
}
