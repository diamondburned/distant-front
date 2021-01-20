package frontend

import (
	"context"
	"html/template"
	"net/http"

	"github.com/diamondburned/distant-front/internal/tmplutil"
	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/diamondburned/distant-front/lib/distance/markup"
)

//go:generate go run github.com/phogolabs/parcello/cmd/parcello -r -i *.go

var Templater = tmplutil.Templater{
	Includes: map[string]string{
		"header": "components/header.html",
		"footer": "components/footer.html",
	},
	Functions: template.FuncMap{
		"markup": func(input string) template.HTML {
			return template.HTML(markup.ToHTML(input))
		},
	},
}

type ctxTypes uint8

const (
	renderStateCtx ctxTypes = iota
)

type RenderState struct {
	Client   *distance.Client
	Observer *distance.Observer
	SiteName string
}

// InjectRenderState injects the render state.
func InjectRenderState(state RenderState) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), renderStateCtx, state)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRenderState gets the render state inside the given context.
func GetRenderState(ctx context.Context) RenderState {
	renderState, ok := ctx.Value(renderStateCtx).(RenderState)
	if !ok {
		panic("no render state in context")
	}
	return renderState
}
