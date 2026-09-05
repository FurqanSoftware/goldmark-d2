package d2

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/d2lang/d2/d2graph"
	"github.com/d2lang/d2/d2layouts/d2dagrelayout"
	"github.com/d2lang/d2/d2layouts/d2elklayout"
	"github.com/d2lang/d2/d2lib"
	"github.com/d2lang/d2/d2renderers/d2svg"
	"github.com/d2lang/d2/d2themes/d2themescatalog"
	"github.com/d2lang/d2/lib/log"
	"github.com/d2lang/d2/lib/textmeasure"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func ptr[T any](v T) *T {
	return &v
}

type HTMLRenderer struct {
	Layout  d2graph.LayoutGraph
	ThemeID *int64
	Sketch  bool
}

func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindBlock, r.Render)
}

func (r *HTMLRenderer) Render(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*Block)
	if !entering {
		w.WriteString("</div>")
		return ast.WalkContinue, nil
	}
	w.WriteString(`<div class="d2">`)

	b := bytes.Buffer{}
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		b.Write(line.Value(src))
	}

	if b.Len() == 0 {
		return ast.WalkContinue, nil
	}

	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return ast.WalkStop, err
	}

	compileOpts := &d2lib.CompileOptions{
		Ruler: ruler,
		LayoutResolver: func(engine string) (d2graph.LayoutGraph, error) {
			if r.Layout != nil {
				return r.Layout, nil
			}
			return layoutForEngine(engine), nil
		},
	}

	renderOpts := &d2svg.RenderOpts{
		Pad:    ptr(int64(d2svg.DEFAULT_PADDING)),
		Sketch: &r.Sketch,
	}
	if r.ThemeID != nil {
		renderOpts.ThemeID = r.ThemeID
	} else {
		renderOpts.ThemeID = &d2themescatalog.CoolClassics.ID
	}

	ctx := context.Background()
	ctx = log.With(ctx, slog.New(slog.DiscardHandler))

	diagram, _, err := d2lib.Compile(ctx, b.String(), compileOpts, renderOpts)
	if err != nil {
		return writeSource(w, b.Bytes())
	}
	out, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return writeSource(w, b.Bytes())
	}

	_, err = w.Write(out)
	return ast.WalkContinue, err
}

// layoutForEngine maps the layout engine D2 resolved for a block, either from
// its d2-config or from D2's own default, onto a layout function.
func layoutForEngine(engine string) d2graph.LayoutGraph {
	switch engine {
	case "elk":
		return d2elklayout.DefaultLayout
	default:
		return d2dagrelayout.DefaultLayout
	}
}

// writeSource writes src as escaped code. It is the fallback for a block that
// cannot be compiled or rendered, so the source must not reach the output
// unescaped.
func writeSource(w util.BufWriter, src []byte) (ast.WalkStatus, error) {
	if _, err := w.WriteString("<pre><code>"); err != nil {
		return ast.WalkContinue, err
	}
	if _, err := w.Write(util.EscapeHTML(src)); err != nil {
		return ast.WalkContinue, err
	}
	_, err := w.WriteString("</code></pre>")
	return ast.WalkContinue, err
}
