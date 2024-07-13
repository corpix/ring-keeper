package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	human "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var dryRun bool

type WatcherConfig struct {
	Directory string  `json:"directory"`
	MaxAmount *int    `json:"max-amount"`
	MaxSize   *string `json:"max-size"`
	maxSize   uint64
	Delay     *int `json:"delay"`
}
type Config []*WatcherConfig

func app() *cli.App {
	app := cli.NewApp()
	app.Description = "keeps specified amount (and/or size) of objects in directory with ring buffer simantics"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "configuration file path",
			Value:   "config.json",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "do not delete anything",
			Value: false,
		},
	}
	app.Action = run
	return app
}

func config(path string) (Config, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var cfg Config
	err = json.NewDecoder(fd).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg) == 0 {
		return nil, errors.New("no watchers are configured")
	}

	for _, item := range cfg {
		stat, err := os.Stat(item.Directory)
		if err != nil {
			return nil, err
		}
		if !stat.IsDir() {
			return nil, errors.Errorf("path %q is not a directory", item.Directory)
		}
		if item.MaxSize != nil {
			item.maxSize, err = human.ParseBytes(*item.MaxSize)
			if err != nil {
				return nil, err
			}
		}
		if item.MaxAmount == nil && item.MaxSize == nil {
			return nil, errors.Errorf(
				"either max amount or max size should be specified for directory %q",
				item.Directory,
			)
		}
		if item.Delay == nil {
			delay := 5 * 60 // 5 minutes
			item.Delay = &delay
		}
	}
	return cfg, nil
}

type FileInfo struct {
	os.FileInfo
	Path string
}

func getMaxAmountOffenders(directory string, maxAmount int) ([]FileInfo, error) {
	var files []FileInfo
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, FileInfo{Path: path, FileInfo: info})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) <= maxAmount {
		return nil, nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	return files[maxAmount:], nil
}

func getMaxSizeOffenders(directory string, maxSize int64) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, FileInfo{Path: path, FileInfo: info})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	skip := -1
	for n, file := range files {
		if maxSize < 0 {
			skip = n
			break
		} else {
			maxSize -= file.Size()
		}
	}
	if skip < 0 {
		return nil, nil
	}
	return files[skip:], nil
}

func runWatcher(wg *sync.WaitGroup, cfg *WatcherConfig) {
	var (
		err                error
		maxAmountOffenders []FileInfo
		maxSizeOffenders   []FileInfo
	)
	defer wg.Done()

	for {
		if cfg.MaxAmount != nil {
			maxAmountOffenders, err = getMaxAmountOffenders(cfg.Directory, *cfg.MaxAmount)
			if err != nil {
				log.Error().Err(err).Msgf("failed to log max amount offenders for %q", cfg.Directory)
				goto delay
			}
		}
		if cfg.MaxSize != nil {
			maxSizeOffenders, err = getMaxSizeOffenders(cfg.Directory, int64(cfg.maxSize))
			if err != nil {
				log.Error().Err(err).Msgf("failed to log max size offenders for %q", cfg.Directory)
				goto delay
			}
		}

		if dryRun {
			fmt.Println("max amount offenders:")
			for _, file := range maxAmountOffenders {
				fmt.Println(file.Path, file.ModTime())
			}
			fmt.Println("")

			fmt.Println("max size offenders:")
			for _, file := range maxSizeOffenders {
				fmt.Println(file.Path, file.ModTime())
			}
		}
	delay:
		if dryRun {
			return
		}
		time.Sleep(time.Second * time.Duration(*cfg.Delay))
	}
}

func run(ctx *cli.Context) error {
	dryRun = ctx.Bool("dry-run")

	cfg, err := config(ctx.String("config"))
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	for _, watcherCfg := range cfg {
		wg.Add(1)
		go runWatcher(wg, watcherCfg)
	}
	wg.Wait()

	return nil
}

func main() {
	app().RunAndExitOnError()
}
