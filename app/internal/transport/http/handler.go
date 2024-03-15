package transport

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/eugene-static/Level0/app/internal/cache"
)

const indexPath = "app/internal/transport/http/ui/index.html"

type Service interface {
	Get(string) ([]byte, error)
}

type Handler struct {
	l       *slog.Logger
	service Service
}

func New(l *slog.Logger, service Service) *Handler {
	return &Handler{
		l:       l,
		service: service,
	}
}

func (h *Handler) Register(router *http.ServeMux) {
	router.HandleFunc("/", h.Index)
	router.HandleFunc("/order", h.Order)
	//fileServer := http.FileServer(http.Dir(staticPath))
	//staticStorage := fmt.Sprintf("/%s/", h.cfg.Server.StaticStorage)
	//router.Handle(staticStorage, http.StripPrefix(staticStorage, fileServer))
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tmpl, err := template.ParseFiles(indexPath)
		if err != nil {
			h.serverError(w, "failed to parse html template", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		if err = tmpl.Execute(w, nil); err != nil {
			h.serverError(w, "failed to apply html template", err)
			return
		}
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			h.serverError(w, "failed to parse form", err)
			return
		}
		key := r.PostFormValue("uid")
		if key == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/order?uid="+key, http.StatusSeeOther)
	default:
		h.clientError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (h *Handler) Order(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.clientError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("uid")
	if key == "" {
		w.Header().Set("Content-Type", "text/html")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data, err := h.service.Get(key)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			h.clientError(w, fmt.Sprintf("key '%s' not found", key), http.StatusNotFound)
		} else {
			h.serverError(w, "failed to get data", err)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		h.serverError(w, "failed to write data", err)
		return
	}

}

func (h *Handler) serverError(w http.ResponseWriter, msg string, err error) {
	h.l.Error(msg, slog.Any("details", err))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (h *Handler) clientError(w http.ResponseWriter, msg string, code int) {
	h.l.Error(msg, slog.String("details", http.StatusText(code)))
	http.Error(w, http.StatusText(code), code)
}
