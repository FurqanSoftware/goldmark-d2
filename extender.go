package d2

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	"oss.terrastruct.com/d2/d2graph"
)

type Extender struct {
	LayoutResolver func(engines string) (d2graph.LayoutGraph, error)
	ThemeID        *int64
	Sketch         bool
}

func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&Transformer{}, 100),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&HTMLRenderer{
			LayoutResolver: e.LayoutResolver,
			ThemeID:        e.ThemeID,
			Sketch:         e.Sketch,
		}, 0),
	))
}
