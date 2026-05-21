package domain

import domain "shelfscan-api/internal/domain/sub"

type AnalysisResponse struct {
	Found        []domain.ProductResult `json:"found"`
	Missing      []domain.MissingItem   `json:"missing"`
	ImageResults []domain.ImageResult   `json:"imageResults"`
}
