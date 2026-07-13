package domain

type AnalysisResponse struct {
	Found        []ProductResult `json:"found"`
	Missing      []MissingItem   `json:"missing"`
	ImageResults []ImageResult   `json:"imageResults"`
}

