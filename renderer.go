package d2

import (
	"bytes"
	"context"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

func ptr[T any](v T) *T {
	return &v
}

type HTMLRenderer struct {
	LayoutResolver func(engines string) (d2graph.LayoutGraph, error)
	ThemeID        *int64
	Sketch         bool
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
	}
	if r.LayoutResolver != nil {
		compileOpts.LayoutResolver = r.LayoutResolver
	} else {
		compileOpts.LayoutResolver = func(engine string) (d2graph.LayoutGraph, error) { return d2dagrelayout.DefaultLayout, nil }
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

	diagram, _, err := d2lib.Compile(context.Background(), b.String(), compileOpts, renderOpts)
	if err != nil {
		_, err = w.Write(b.Bytes())
		return ast.WalkContinue, err
	}
	out, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		_, err = w.Write(b.Bytes())
		return ast.WalkContinue, err
	}

	_, err = w.Write(out)
	return ast.WalkContinue, err
}
