package domain

type ProductResult struct {
	Name     string  `json:"name"`
	Sku      string  `json:"sku"`
	ImageURL string  `json:"imageUrl"`
	Score    float32 `json:"score"`
	Count    int     `json:"count"`
}
