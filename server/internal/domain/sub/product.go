package domain

type ProductResult struct {
	Name     string  `json:"name"`
	ImageURL string  `json:"imageUrl"`
	Score    float32 `json:"score"`
}
