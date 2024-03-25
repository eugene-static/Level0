package transport

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/eugene-static/Level0/app/internal/cache"
	"github.com/eugene-static/Level0/app/lib/config"
)

const (
	ctText = "text/html"
	ctJson = "application/json"
	ctForm = "application/x-www-form-urlencoded"
	ct     = "Content-Type"
)

type Service interface {
	Get(string) ([]byte, error)
}

type Handler struct {
	l       *slog.Logger
	cfg     *config.Server
	service Service
}

func New(l *slog.Logger, cfg *config.Server, service Service) *Handler {
	return &Handler{
		l:       l,
		cfg:     cfg,
		service: service,
	}
}

func (h *Handler) Register(router *http.ServeMux) {
	fileServer := http.FileServer(http.Dir(h.cfg.StaticPath))
	staticStorage := fmt.Sprintf("/%s/", "static")
	router.Handle(staticStorage, http.StripPrefix(staticStorage, fileServer))
	router.Handle("/", h.index(h.cfg.TemplatePath))
	router.Handle("/order", h.order())
}

func (h *Handler) index(templatePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		switch r.Method {
		case http.MethodGet:
			tmpl, err := template.ParseFiles(templatePath)
			if err != nil {
				h.serverError(w, "failed to parse html template", err)
				return
			}
			w.Header().Set(ct, ctText)
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
			key := r.FormValue("uid")
			http.Redirect(w, r, "/order?uid="+key, http.StatusSeeOther)
		default:
			h.clientError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func (h *Handler) order() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			h.clientError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		key := r.URL.Query().Get("uid")
		data, err := h.service.Get(key)
		h.l.Info("getting data", slog.String("uid", key))
		if err != nil {
			if errors.Is(err, cache.ErrNotFound) {
				h.notFound(w, r)
			} else {
				h.serverError(w, "failed to get data", err)
			}
			return
		}
		h.l.Info("data found", slog.String("uid", key))
		w.Header().Set(ct, ctJson)
		w.WriteHeader(http.StatusFound)
		_, err = w.Write(data)
		if err != nil {
			h.serverError(w, "failed to write data", err)
			return
		}
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

func (h *Handler) notFound(w http.ResponseWriter, r *http.Request) {
	h.l.Warn("not found", slog.String("details", http.StatusText(http.StatusNotFound)))
	http.NotFound(w, r)
}
