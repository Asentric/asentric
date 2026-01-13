package dispatcher

import "github.com/asentric/asentric/pkg/asentric"

// ContextBuilderAdapter adapts DefaultContextBuilder to the ContextBuilder interface.
// This provides a clean way to use DefaultContextBuilder with the dispatcher.
type ContextBuilderAdapter struct {
	builder *DefaultContextBuilder
}

// NewContextBuilderAdapter creates a new adapter wrapping DefaultContextBuilder.
func NewContextBuilderAdapter(builder *DefaultContextBuilder) *ContextBuilderAdapter {
	return &ContextBuilderAdapter{builder: builder}
}

// Build implements the ContextBuilder interface.
func (a *ContextBuilderAdapter) Build(event asentric.Event) asentric.Context {
	ctx, _ := a.builder.Build(event)
	return ctx
}

// Ensure ContextBuilderAdapter implements ContextBuilder
var _ ContextBuilder = (*ContextBuilderAdapter)(nil)
