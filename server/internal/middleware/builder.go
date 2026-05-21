package middleware

type Middleware struct {
	allowedOrigins string
	allowedMethods string
	allowedHeaders string
}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) WithAllowedOrigins(origins string) *Middleware {
	m.allowedOrigins = origins
	return m
}

func (m *Middleware) WithAllowedMethods(methods string) *Middleware {
	m.allowedMethods = methods
	return m
}

func (m *Middleware) WithAllowedHeaders(headers string) *Middleware {
	m.allowedHeaders = headers
	return m
}
