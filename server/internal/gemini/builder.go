package gemini

type GeminiClient struct {
	GenerativeModel string
	Apikey          string
	Prompt          string

	Temperature     int
	MaxOutputTokens int
}

func NewGeminiClient() *GeminiClient {
	return &GeminiClient{}
}

func (g *GeminiClient) WithGenerativeModel(model string) *GeminiClient {
	g.GenerativeModel = model
	return g
}

func (g *GeminiClient) WithApiKey(key string) *GeminiClient {
	g.Apikey = key
	return g

}

func (g *GeminiClient) WithTemperature(temp int) *GeminiClient {
	g.Temperature = temp
	return g
}

func (g *GeminiClient) WithMaxOutputTokens(tokens int) *GeminiClient {
	g.MaxOutputTokens = tokens
	return g

}

func (g *GeminiClient) WithPrompt(prompt string) *GeminiClient {
	g.Prompt = prompt
	return g
}
