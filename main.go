package main

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailshandler "github.com/yukihito-jokyu/context-lab/internal/handler/wails"
	"github.com/yukihito-jokyu/context-lab/internal/logger"
	"github.com/yukihito-jokyu/context-lab/internal/repository/acp"
	"github.com/yukihito-jokyu/context-lab/internal/repository/sqlite"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
)

const applicationDirectoryName = "context-lab"

//go:embed all:frontend/dist
var assets embed.FS

// アプリケーション起動
func main() {
	app := NewApp()
	appLogger := logger.New(slog.LevelInfo)
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		appLogger.Error(context.Background(), "find user config directory", err)

		return
	}

	store, err := sqlite.Open(filepath.Join(configDirectory, applicationDirectoryName))
	if err != nil {
		appLogger.Error(context.Background(), "initialize experiment store", err)

		return
	}
	defer func() {
		if err := store.Close(); err != nil {
			appLogger.Error(context.Background(), "close experiment store", err)
		}
	}()

	experimentsHandler := wailshandler.NewExperimentsHandler(usecase.NewListExperiments(store), appLogger)
	experimentPreparationsHandler := wailshandler.NewExperimentPreparationsHandlerWithConditions(
		usecase.NewGetExperimentPreparation(store),
		appLogger,
		usecase.NewSaveExperimentPreparationDraft(store),
		usecase.NewFixExperimentConditions(store),
	)
	preparationsHandler := wailshandler.NewPreparationsHandler(usecase.NewListPreparations(store), appLogger)
	briefingAdapter := acp.NewCodexBriefingAdapter(filepath.Join(configDirectory, applicationDirectoryName))
	experimentBriefingsHandler := wailshandler.NewExperimentBriefingsHandler(
		usecase.NewStartExperimentBriefing(store, briefingAdapter),
		usecase.NewSendExperimentBriefMessage(store, briefingAdapter),
		usecase.NewGetExperimentBriefing(store),
		usecase.NewCreateExperimentFromBrief(store),
		appLogger,
		usecase.NewStopExperimentBriefing(store, briefingAdapter),
	)

	err = wails.Run(&options.App{
		Title:       "Context Lab",
		Width:       1280,
		Height:      800,
		AssetServer: &assetserver.Options{Assets: assets},
		OnStartup:   app.startup,
		Bind:        []interface{}{experimentsHandler, experimentPreparationsHandler, preparationsHandler, experimentBriefingsHandler},
	})
	if err != nil {
		appLogger.Error(context.Background(), "run application", err)
	}
}
