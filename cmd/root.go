package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/treethought/ethtop/config"
	"github.com/treethought/ethtop/ui"
)

var (
	configPath = os.Getenv("CONFIG_FILE")
	cfg        *config.Config
	logFile    *os.File
)

var rootCmd = &cobra.Command{
	Use:   "ethtop",
	Short: "terminal based evm network monitor",
	Run: func(cmd *cobra.Command, args []string) {
		defer logFile.Close()
		appCtx := &ui.AppContext{}
		app := ui.NewApp(cfg, appCtx)
		p := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file (default is $HOME/.ethtop/config.yaml)")
}

func initConfig() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		return
	}
	var err error
	if configPath == "" {
		if _, err := os.Stat("config.yaml"); err == nil {
			configPath = "config.yaml"
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatal("failed to find default config file: ", err)
			}
			configPath = filepath.Join(homeDir, ".ethtop", "config.yaml")
		}
	}
	cfg, err = config.ReadConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatal("failed to fing config file, run `ethtop init` to create one")
		}
		log.Fatal("failed to read config: ", err)
	}

	lf := "ethtop.log"
	dir := filepath.Dir(lf)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	logFile, err = tea.LogToFile(lf, "debug")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.SetOutput(logFile)
	log.Println("loaded config: ", configPath)
}
