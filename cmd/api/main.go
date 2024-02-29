package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofor-little/env"
	_ "github.com/lib/pq"
	"greenlight.vysotsky.com/internal/data"
)

const version = "1.0.0"

type config struct {
	port int
	env string
	db struct {
		dsn string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime string
	}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
}


func main() {
	var conf config
	flag.IntVar(&conf.port, "port", 4000, "Api server port")
	flag.StringVar(&conf.env, "env", "development", "Environment (development|staging|pruduction)")

	if err := env.Load(".env"); err != nil {
		// panic(err)
	}

	default_dsn, _ := env.MustGet("DB_DSN")
	flag.StringVar(&conf.db.dsn, "db-dsn", default_dsn, "PostgreSQL DSN")
	flag.IntVar(&conf.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&conf.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&conf.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle time (10s|30m)")
	flag.Parse()

	if len(conf.db.dsn) == 0 {
		panic("database dsn was not provided neither in .env file nor as a -db-dsn flag")
	}

	fmt.Println("port:", conf.port)
	logger := log.New(os.Stdout, "", log.Ldate | log.Ltime)

	db, err := openDB(conf)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Printf("database connection established")

	app := &application{
		config: conf,
		logger: logger,
		models: data.NewModels(db),
	}

	router := app.routes()
	handler := app.addLoggging(router)

	server := &http.Server {
		// Addr: fmt.Sprintf(":%d", conf.port),
		Addr: fmt.Sprintf("localhost:%d", conf.port),
		Handler: handler,
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	logger.Printf("starting %s server on %s", conf.env, server.Addr)
	err = server.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime) 
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getLogger() *log.Logger {
	file, err := os.OpenFile("log.txt", os.O_RDWR | os.O_APPEND | os.O_CREATE, 0666) 
	if err != nil {
		fmt.Println("error crating logger file!")
		file = os.Stdout
	}
	writer := bufio.NewWriter(file)
	i, err := fmt.Fprintln(writer, "hello bichezz")
	defer writer.Flush()
	defer file.Close()
	fmt.Println("i = ", i)
	fmt.Println("err1 = ", err)
	logger := log.New(writer, "", log.Ldate | log.Ltime)
	logger.SetOutput(writer)
	return logger
}
