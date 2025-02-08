package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var configFilePath string

func init() {
	flag.StringVar(&configFilePath, "config", "config.yaml", "Config file name")
}

func main() {
	flag.Parse()

	log, err := initLogger()
	if err != nil {
		panic(err)
	}

	cfg, err := initConfig(configFilePath)
	if err != nil {
		log.Fatal("initConfig", zap.Error(err))
	}
	log.Info("Config", zap.String("file", configFilePath), zap.Any("content", cfg))

	sender, err := NewFeishuRobot(log, cfg.Alert.NotifyUrl)
	if err != nil {
		log.Fatal("NewFeishuRobot", zap.Error(err))
	}

	app, stop := NewApp(log, cfg, sender)
	handleShutdown(stop)
	app.Run()
}

func handleShutdown(f func()) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		f()
	}()
}

func initLogger() (*zap.Logger, error) {
	log, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		return nil, err
	}

	return log, nil
}

func initConfig(filePath string) (*Config, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "read config file")
	}

	var cfg Config
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config")
	}

	if err = cfg.Complete(); err != nil {
		return nil, errors.Wrap(err, "complete config")
	}

	return &cfg, nil
}
