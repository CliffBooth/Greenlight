package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	router := app.routes()
	handler := app.recoverPanic(router)
	handler = app.rateLimit(handler)
	handler = app.logRequests(handler)

	server := &http.Server {
		// Addr: fmt.Sprintf(":%d", conf.port),
		Addr: fmt.Sprintf("localhost:%d", app.config.port),
		Handler: handler,
		ErrorLog: log.New(app.logger, "", 0),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()
		shutdownError <- server.Shutdown(ctx)
	}()

	app.logger.PrintInfo("starting server", map[string]string {
		"addr": server.Addr,
		"env": app.config.env,
	})

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdownError
	if err != nil {
		return err
	}
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": server.Addr,
	})
	return nil
}