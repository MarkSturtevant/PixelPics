package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/coder/websocket"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func handleWS(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithCancelCause(r.Context())

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return fmt.Errorf("Error during websocket.Accept() %w\n", err)
	}

	var innerErr error
	defer func() {
		if tempErr := conn.CloseNow(); tempErr != nil {
			innerErr = fmt.Errorf("Error during conn.CloseNow() %w\n", tempErr)
		}
	}()

	defer cancel(errors.New("cancel websocket closed"))

	errCh := make(chan error, 1)

	textResChan := make(chan string)

	go func() {
		defer cancel(errors.New("cancel websocket read closed"))

		for {
			msgType, data, err := conn.Read(ctx)
			if err != nil {
				errCh <- fmt.Errorf("Error during conn.Read() %w\n", err)
				return
			}

			if msgType == websocket.MessageBinary {
				slog.Info(fmt.Sprintf("Received binary message: %v\n", data))
			} else if msgType == websocket.MessageText {
				slog.Info(fmt.Sprintf("Received text message: %v\n", data))
			} else {
				errCh <- fmt.Errorf("Error during conn.Read() invalid message type: %v\n", msgType.String())
				return
			}
		}
	}()

	for {
		select {
		case textRes := <-textResChan:
			err := conn.Write(ctx, websocket.MessageText, []byte(textRes))
			if err != nil {
				errCh <- fmt.Errorf("Error during conn.Write() %w\n", err)
				continue
			}
		case err := <-errCh:
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				slog.Info("Closing websocket connection", "code", closeErr.Code.String(), "reason", closeErr.Reason)
				return innerErr
			} else {
				return fmt.Errorf("Error during websocket endpoint: %w\n", err)
			}

		}
	}
}

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

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

		se.Router.GET("/api/v1/ws", func(e *core.RequestEvent) error {
			err := handleWS(e.Response, e.Request)
			if err != nil {
				slog.Error(err.Error())
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
