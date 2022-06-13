package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	// postgres drivers
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "8080"

var app *Config

func main() {
	// connect to the database
	conn := initDB()

	// create sessions
	session := initSession()

	// create channels

	// create waitgroup
	wg := sync.WaitGroup{}

	// set up the application config
	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime|log.Ldate)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ltime|log.Ldate|log.Lshortfile)

	app = &Config{
		DB:       conn,
		Session:  session,
		Wait:     &wg,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
	}

	// set up mail

	// listen for web connections
	app.serve()
}

func initDB() *sql.DB {
	conn := connectToDB()
	if conn == nil {
		log.Panic("could not connect to DB")
	}
	return conn
}

func connectToDB() *sql.DB {
	count := 0

	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			count++
			log.Printf("error: %v\n", err)
			log.Println("backing off from db...")
			time.Sleep(2 * time.Second)
		} else {
			log.Println("connected to DB")
			return connection
		}

		if count > 10 {
			return nil
		}
	}

}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// does it respond?
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initSession() *scs.SessionManager {
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}

func initRedis() *redis.Pool {
	redisPool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}

	return redisPool
}

func (app *Config) serve() {
	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	app.InfoLog.Printf("starting server on port %s\n", webPort)
	err := srv.ListenAndServe()
	if err != nil {
		log.Panicln(err)
	}
}
