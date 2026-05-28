package domain

type MissingItem struct {
	Name     string `json:"name"`
	Sku      string `json:"sku"`
	ImageURL string `json:"imageUrl"`
}
