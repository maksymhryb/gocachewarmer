package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maksymhryb/gocachewarmer/config"
	p "github.com/maksymhryb/gocachewarmer/pipeline"
	"github.com/maksymhryb/gocachewarmer/ui"
	"gopkg.in/yaml.v3"
)

func main() {
	config := config.InitCliConfig()
	initLog(config)

	switch flag.Arg(0) {
	case "run":
		run(config)
	case "generate-config":
		generateConfig()
	default:
		flag.Usage()
	}
}

func run(config *config.Config) {
	if config.SitemapUrl == "" {
		fmt.Println("Sitemap URL is not specified")
		os.Exit(1)
	}

	getCtx, genCancel := context.WithCancel(context.Background())
	defer genCancel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stageLoadSitemap := p.LoadSitemap(getCtx, config)
	stageFilterInput := p.FilterInput(ctx, config, stageLoadSitemap)
	stageLimitInput := p.LimitInput(ctx, genCancel, config, stageFilterInput)
	stageLimitInputTeeN := p.TeeN(ctx, config, stageLimitInput, 2, 10000)
	stageCount := p.Count(ctx, config, stageLimitInputTeeN[0], 100)
	stageWarmup := p.Warmup(ctx, config, stageLimitInputTeeN[1])
	stageWarmupTeeN := p.TeeN(ctx, config, stageWarmup, 3, 0)
	stageAggregate := p.Aggregate(ctx, config, stageWarmupTeeN[1])
	stageFilterOutput := p.FilterOutput(ctx, config, stageWarmupTeeN[2])
	stageSaveReport := p.SaveReport(ctx, config, stageFilterOutput)
	stageAggregateTeeN := p.TeeN(ctx, config, stageAggregate, 2, 0)
	stageSaveAggregatedReport := p.SaveAggregatedReport(ctx, config, stageAggregateTeeN[1])

	p := tea.NewProgram(ui.InitialModel(config.DryRun), tea.WithAltScreen())

	go func() {
		for val := range stageWarmupTeeN[0] {
			p.Send(val)
		}
	}()

	go func() {
		for val := range stageAggregateTeeN[0] {
			p.Send(val)
		}
	}()

	go func() {
		for val := range stageCount {
			p.Send(ui.TotalCounterMsg(val))
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}

	<-stageSaveReport
	<-stageSaveAggregatedReport
}

func generateConfig() {
	_, err := os.Stat("config.yaml")

	if err == nil {
		var input string
		fmt.Print("config.yaml file already exists, do you want to override it? [Y/n]: ")
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))
		if !(input == "y" || input == "yes" || input == "") {
			return
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	data, err := yaml.Marshal(config.DefaultConfig)
	if err != nil {

	}
	err = os.WriteFile("config.yaml", data, 0o644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done.")
}

func initLog(config *config.Config) {
	if config.SkipLogs {
		log.SetOutput(io.Discard)
	} else {
		file, err := os.OpenFile(config.LogName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Cannot create log file \"%s\", error: %s", config.LogName, err)
		}
		log.SetOutput(file)
	}
}
