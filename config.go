package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

type Config struct {
	Domains []string `yaml:"domains" json:"domains"`
	Alert   struct {
		NotifyUrl        string        `yaml:"notifyUrl" json:"notifyUrl"`
		BeforeExpiredStr string        `yaml:"beforeExpired" json:"beforeExpired"`
		BeforeExpired    time.Duration `yaml:"-" json:"-"`
	} `yaml:"alert" json:"alert"`
}

func (c *Config) Complete() error {
	if len(c.Domains) == 0 {
		return errors.New("no any domain configured")
	}

	beforeExpired, err := ParseDuration(c.Alert.BeforeExpiredStr)
	if err != nil {
		return fmt.Errorf("invalid beforeExpired: %w", err)
	}

	if beforeExpired.Hours() > 30*24 || beforeExpired.Hours() < 3*24 {
		return errors.New("invalid beforeExpired: should be between 3 and 30 days")
	}

	_, err = url.ParseRequestURI(c.Alert.NotifyUrl)
	if err != nil {
		return fmt.Errorf("invalid notifyUrl: %w", err)
	}

	c.Alert.BeforeExpired = beforeExpired
	return nil
}
