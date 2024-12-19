package api

import (
	"fmt"
	"log/slog"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/MarkSturtevant/PixelPics/backend/api/wshandler"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterRoutes(se *core.ServeEvent) error {
	frontendURL, err := url.Parse("http://localhost:5173")
	if err != nil {
		return fmt.Errorf("Error during Parse() %s: %w\n", "http://localhost:5173", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(frontendURL)

	se.Router.BindFunc(func(e *core.RequestEvent) error {
		if !strings.HasPrefix(e.Request.URL.Path, "/api") &&
			!strings.HasPrefix(e.Request.URL.Path, "/_") {
			proxy.ServeHTTP(e.Response, e.Request)
			return nil
		}
		return e.Next()
	})

	se.Router.GET("/api/v1/ws", func(e *core.RequestEvent) error {
		err := wshandler.HandleWS(e.Response, e.Request)
		if err != nil {
			slog.Error(err.Error())
		}

		return e.Next()
	})

	return nil
}
