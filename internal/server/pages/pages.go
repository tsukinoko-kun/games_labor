package pages

import (
	"gameslabor/internal/server/context"
	"log"
	"net/http"

	"github.com/a-h/templ"
)

var pageRegister = map[string]func() templ.Component{}

func Handler(w http.ResponseWriter, r *http.Request) {
	if h, ok := pageRegister[r.URL.Path]; ok {
		if err := h().Render(context.From(w, r), w); err != nil {
			log.Println("Error in pages handler", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if h, ok := pageRegister[r.URL.Path+"/"]; ok {
		if err := h().Render(context.From(w, r), w); err != nil {
			log.Println("Error in pages handler", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}
