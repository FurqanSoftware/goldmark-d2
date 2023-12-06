package d2

import (
	"bytes"
	"context"
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

type HTMLRenderer struct {
	Extender
}

func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindBlock, r.Render)
}

func (r *HTMLRenderer) compileOptions() (*d2lib.CompileOptions, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, err
	}

	ret := &d2lib.CompileOptions{
		Ruler:  ruler,
		Layout: r.Layout,
		LayoutResolver: func(engine string) (d2graph.LayoutGraph, error) {
			if engine == "" || engine == "dagre" {
				return d2dagrelayout.DefaultLayout, nil
			} else if engine == "elk" {
				return d2elklayout.DefaultLayout, nil
			}
			return nil, fmt.Errorf("unknown engine '%s'", engine)
		},
	}
	return ret, nil
}

func (r *HTMLRenderer) renderOptions() *d2svg.RenderOpts {
	renderOpts := &d2svg.RenderOpts{
		Pad:    Ptr(int64(d2svg.DEFAULT_PADDING)),
		Sketch: r.Sketch,
	}
	if r.ThemeID != nil {
		renderOpts.ThemeID = r.ThemeID
	} else {
		renderOpts.ThemeID = &d2themescatalog.CoolClassics.ID
	}

	return renderOpts
}

func (r *HTMLRenderer) Render(
	w util.BufWriter,
	src []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	n := node.(*Block)
	if !entering {
		if _, err := w.WriteString("</div>"); err != nil {
			return ast.WalkStop, err
		}
		return ast.WalkContinue, nil
	}

	if _, err := w.WriteString(`<div class="d2">`); err != nil {
		return ast.WalkStop, err

	}

	b := bytes.Buffer{}
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		b.Write(line.Value(src))
	}

	if b.Len() == 0 {
		return ast.WalkContinue, nil
	}

	compileOpts, err := r.compileOptions()
	if err != nil {
		return ast.WalkStop, err
	}

	renderOpts := r.renderOptions()
	ctx := log.Stderr(context.Background())
	diagram, _, err := d2lib.Compile(ctx, b.String(), compileOpts, renderOpts)
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
