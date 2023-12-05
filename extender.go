package d2

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func Ptr[T any](v T) *T {
	return &v
}

type Extender struct {
	Layout  *string `json:"layout,omitempty"`
	ThemeID *int64  `json:"theme_id,omitempty"`
	Sketch  *bool   `json:"sketch,omitempty"`
}

func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&Transformer{}, 100),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&HTMLRenderer{*e}, 0),
	))
}
