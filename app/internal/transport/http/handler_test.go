package transport

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/eugene-static/Level0/app/internal/cache"
	"github.com/eugene-static/Level0/app/internal/service"
	"github.com/eugene-static/Level0/app/internal/storage/postgres"
	"github.com/eugene-static/Level0/app/lib/config"
	"github.com/eugene-static/Level0/app/lib/logger"
)

func TestHandler(t *testing.T) {
	cases := []struct {
		url    string
		method string
		body   string
		header string
		want   int
	}{
		{
			url:    "/",
			method: http.MethodGet,
			body:   "",
			header: ctText,
			want:   http.StatusOK,
		},
		{
			url:    "/test",
			method: http.MethodGet,
			body:   "",
			header: ctText,
			want:   http.StatusSeeOther,
		},
		{
			url:    "/",
			method: http.MethodPost,
			body:   "test",
			header: ctForm,
			want:   http.StatusSeeOther,
		},
		{
			url:    "/",
			method: http.MethodPut,
			body:   "",
			header: ctText,
			want:   http.StatusMethodNotAllowed,
		},
		{
			url:    "/order?uid=test",
			method: http.MethodGet,
			body:   "",
			header: ctJson,
			want:   http.StatusFound,
		},
		{
			url:    "/order?uid=test",
			method: http.MethodPost,
			body:   "",
			header: ctJson,
			want:   http.StatusMethodNotAllowed,
		},
	}
	storage := &postgres.Storage{}
	cch := cache.New(&sync.RWMutex{})
	cch.Set("test", []byte(`{"ok": "ok"}`))
	srv := service.New(storage, cch)
	hdlr := New(logger.NewEmpty(), &config.Server{TemplatePath: "./ui/index.html"}, srv)
	router := http.NewServeMux()
	hdlr.Register(router)
	for _, c := range cases {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest(c.method, c.url, strings.NewReader(c.body))
		if err != nil {
			t.Fatal(err)
		}
		if c.header == ctForm {
			req.PostForm = url.Values{"uid": {c.body}}
		}
		rr.Header().Set(ct, c.header)
		router.ServeHTTP(rr, req)
		if status := rr.Result().StatusCode; status != c.want {
			t.Errorf("want %v, got %v", c.want, status)
		} else {
			t.Logf("url: %s, method: %s, want: %d, got: %d", req.URL, req.Method, c.want, status)
		}
	}

}
