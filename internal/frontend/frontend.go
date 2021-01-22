package frontend

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/diamondburned/distant-front/internal/tmplutil"
	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/diamondburned/distant-front/lib/distance/markup"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/phogolabs/parcello"
)

//go:generate go run github.com/phogolabs/parcello/cmd/parcello -r -i *.go

var Templater = tmplutil.Templater{
	Includes: map[string]string{
		"css":        "components/css.html",
		"header":     "components/header.html",
		"footer":     "components/footer.html",
		"empty-card": "components/empty-card.html",
	},
	Functions: template.FuncMap{
		"markup": func(input string) template.HTML {
			return template.HTML(markup.ToHTML(input))
		},
		"rgbaHex": func(rgba [4]float32) string {
			color := colorful.FastLinearRgb(
				float64(rgba[0]),
				float64(rgba[1]),
				float64(rgba[2]),
			)
			return color.Hex()
		},
	},
}

type ctxTypes uint8

const (
	renderStateCtx ctxTypes = iota
)

type RenderState struct {
	Client      *distance.Client
	Observer    *distance.Observer
	SiteName    string
	DistanceURL *url.URL
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

// ExecuteTemplate executes the template with the RenderState.
func ExecuteTemplate(w http.ResponseWriter, r *http.Request, sub *tmplutil.Subtemplate) {
	if err := sub.Execute(w, GetRenderState(r.Context())); err != nil {
		log.Println("Error rendering:", err)
	}
}

// ExecuteNamedTemplate executes the named template with the RenderState.
func ExecuteNamedTemplate(w http.ResponseWriter, r *http.Request, name string) {
	if err := Templater.Execute(w, name, GetRenderState(r.Context())); err != nil {
		log.Println("Error rendering:", err)
	}
}

// MountStatic mounts the static route.
func MountStatic() http.Handler {
	d, err := parcello.Manager.Dir("static/")
	if err != nil {
		log.Fatalln("Static not found:", err)
	}

	return http.StripPrefix("/static", http.FileServer(d))
}
