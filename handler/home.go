package handler

import (
	"dreampicai/view/home"
	"net/http"
	"time"
)

func HandleLongProcess(w http.ResponseWriter, r *http.Request) error {
	time.Sleep(time.Second * 5)
	return render(r, w, home.UserLikes(234436))
}

func HandlerHomeIndex(w http.ResponseWriter, r *http.Request) error {
	return render(r, w, home.Index())
}
