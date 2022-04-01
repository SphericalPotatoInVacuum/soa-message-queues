package serverwaiter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func Wait(ctx context.Context, URL string) error {
	contextLogger := log.WithField("addr", URL)

	contextLogger.Info("Waiting for address")
	u, err := url.Parse(URL)
	if err != nil {
		return err
	}
	port := u.Port()
	if port == "" {
		port = u.Scheme
	}
	addr := fmt.Sprintf("%s:%s", u.Hostname(), port)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return errors.New("Context timed out")
			default:
				break
			}
			conn, err := net.Dial("tcp", addr)
			if err == nil {
				conn.Close()
				return nil
			}
			contextLogger.WithError(err).Info("Target is not up, retrying")
			time.Sleep(time.Duration(1) * time.Second)
		}
	})
	err = g.Wait()
	if err != nil {
		contextLogger.WithError(err).Error("Wait timed out")
	}
	return err
}
