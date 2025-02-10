package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type App struct {
	log    *zap.Logger
	cfg    *Config
	sender *FeishuRobot

	willExpireSoon        map[string]time.Duration // domain:lastTime
	continuousErrorCounts map[string]int           // domain:count

	ctx context.Context
}

func NewApp(log *zap.Logger, cfg *Config, sender *FeishuRobot) (*App, func()) {
	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		ctx:    ctx,
		log:    log,
		cfg:    cfg,
		sender: sender,

		willExpireSoon:        make(map[string]time.Duration),
		continuousErrorCounts: make(map[string]int),
	}

	return app, func() {
		cancel()
		log.Info("Stopped")
	}
}

func (a *App) Run() {
	// Twice a day
	ticker := time.NewTicker(time.Hour * 12)
	defer ticker.Stop()

	a.log.Info("Started")
	for {
		a.checkOnce(a.ctx)
		select {
		case <-ticker.C:
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *App) checkOnce(ctx context.Context) {
	a.log.Debug("Check once")
	for _, domain := range a.cfg.Domains {
		expiredTime, err := parseCertExpiredTime(ctx, domain)
		if err != nil {
			select {
			case <-a.ctx.Done():
				return
			default:
			}
			a.log.Error("parseCertExpiredTime", zap.String("domain", domain), zap.Error(err))
			a.continuousErrorCounts[domain]++
			continue
		}

		// Reset error count
		a.continuousErrorCounts[domain] = 0

		lastTime := time.Until(expiredTime)
		logFields := []zap.Field{zap.String("domain", domain), zap.Int64("days", int64(lastTime.Hours())/24)}
		if lastTime > a.cfg.Alert.BeforeExpired {
			// Health, remove it from alert map
			a.log.Debug("Expiration", logFields...)
			delete(a.willExpireSoon, domain)
			continue
		}

		// Will expire soon
		a.willExpireSoon[domain] = lastTime
		a.log.Info("Expiration", logFields...)
	}

	// Alert
	var message string
	if len(a.willExpireSoon) > 0 {
		message = "HTTPS Certification will expire soon:<br>"
		for domain, lastDuration := range a.willExpireSoon {
			message += fmt.Sprintf("- %s (remaining: %dd)<br>", domain, int64(lastDuration.Hours())/24)
		}
	}

	var errorMessage string
	if len(a.continuousErrorCounts) > 0 {
		for domain, errorCount := range a.continuousErrorCounts {
			if errorCount >= 5 {
				errorMessage += fmt.Sprintf("- %s (errorCount: %d)<br>", domain, errorCount)
			}
		}
		if errorMessage != "" {
			errorMessage = "HTTPS Certification check failed:<br>" + errorMessage
		}
	}
	message = message + errorMessage

	if message != "" {
		err := a.sender.SendWarnMessage(TemplateData{
			Title:   "Certification Error: Fired",
			Content: message,
			Foot:    "certificate-sentry",
		})
		if err != nil {
			a.log.Error("Send message", zap.Error(err))
		}
	}
}

func parseCertExpiredTime(ctx context.Context, domain string) (time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	d := &tls.Dialer{
		Config: &tls.Config{InsecureSkipVerify: true},
	}
	conn, err := d.DialContext(ctx, "tcp", domain+":443")
	if err != nil {
		return time.Time{}, errors.WithStack(err)
	}

	defer conn.Close()

	certs := conn.(*tls.Conn).ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return time.Time{}, errors.New("no tls certification")
	}

	return certs[0].NotAfter, nil
}
