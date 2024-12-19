package main

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"github.com/MarkSturtevant/PixelPics/backend/api"
	_ "github.com/MarkSturtevant/PixelPics/migrations"
)

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		err := api.RegisterRoutes(se)
		if err != nil {
			return fmt.Errorf("Error during RegisterRoutes() %w\n", err)
		}

		err = se.Next()
		if err != nil {
			return fmt.Errorf("Error during se.Next() %w\n", err)
		}

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
