package handler

import (
	"dreampicai/view/home"
	"net/http"
)

func HandlerHomeIndex(w http.ResponseWriter, r *http.Request) error {
	return home.Index().Render(r.Context(), w)
}
