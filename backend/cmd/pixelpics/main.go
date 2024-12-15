package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	// app.OnServe().BindFunc(func(se *core.ServeEvent) error {
	// 	// serves static files from the provided public dir (if exists)
	// 	se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))
	//
	// 	return se.Next()
	// })

	frontendURL, err := url.Parse("http://localhost:5173")
	if err != nil {
		slog.Error(fmt.Errorf("Error during Parse() %s: %w\n", "http://localhost:5173", err).Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(frontendURL)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.BindFunc(func(e *core.RequestEvent) error {
			if !strings.HasPrefix(e.Request.URL.Path, "/api") &&
				!strings.HasPrefix(e.Request.URL.Path, "/_") {
				proxy.ServeHTTP(e.Response, e.Request)
				return nil
			}
			return e.Next()
		})

		err := se.Next()
		if err != nil {
			slog.Error(err.Error())
		}

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
