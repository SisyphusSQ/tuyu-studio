package main

import (
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/SisyphusSQ/tuyu-studio/internal/shell"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp(shell.NewService(time.Now()))

	err := wails.Run(&options.App{
		Title:     "Tuyu Studio",
		Width:     1280,
		Height:    820,
		MinWidth:  1024,
		MinHeight: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 20, G: 24, B: 31, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start Tuyu Studio shell: %v\n", err)
		os.Exit(1)
	}
}
