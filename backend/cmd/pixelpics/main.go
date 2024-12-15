package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func proxy(w http.ResponseWriter, r *http.Request) error {
	newURL, err := url.Parse(r.URL.String())
	if err != nil {
		return fmt.Errorf("Error during Parse() %s: %w\n", r.URL.String(), err)
	}

	newURL.Scheme = "http"
	newURL.Host = "localhost:5173"

	req, err := http.NewRequest(r.Method, newURL.String(), r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("Error during NewRequest() %s: %w\n", r.URL.String(), err)
	}

	// copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("Error during Do() %s: %w\n", r.URL.String(), err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("Error during Copy() %s: %w\n", r.URL.String(), err)
	}

	return nil
}

func main() {
	app := pocketbase.New()

	// app.OnServe().BindFunc(func(se *core.ServeEvent) error {
	// 	// serves static files from the provided public dir (if exists)
	// 	se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))
	//
	// 	return se.Next()
	// })

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.BindFunc(func(e *core.RequestEvent) error {
			if !strings.HasPrefix(e.Request.URL.Path, "/api") &&
				!strings.HasPrefix(e.Request.URL.Path, "/_") {
				err := proxy(e.Response, e.Request)
				if err != nil {
					slog.Error(err.Error())
				}
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
