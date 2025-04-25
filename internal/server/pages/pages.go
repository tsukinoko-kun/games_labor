package pages

import (
	"gameslabor/internal/server/context"
	"net/http"

	"github.com/a-h/templ"
)

var pageRegister = map[string]func() templ.Component{}

func Handler(w http.ResponseWriter, r *http.Request) {
	if h, ok := pageRegister[r.URL.Path]; ok {
		h().Render(context.From(w, r), w)
		return
	}
	if h, ok := pageRegister[r.URL.Path+"/"]; ok {
		h().Render(context.From(w, r), w)
		return
	}
}
