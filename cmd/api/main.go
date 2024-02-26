package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

//p.

const version = "1.0.0"

type config struct {
	port int
	env string
}

type application struct {
	config config
	logger *log.Logger
}


func main() {
	var conf config
	flag.IntVar(&conf.port, "port", 4000, "Api server port")
	flag.StringVar(&conf.env, "env", "development", "Environment (development|staging|pruduction)")
	
	flag.Parse()
	fmt.Println("port:", conf.port)
	logger := log.New(os.Stdout, "", log.Ldate | log.Ltime)

	app := &application{
		config: conf,
		logger: logger,
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
	err := server.ListenAndServe()
	logger.Fatal(err)
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
