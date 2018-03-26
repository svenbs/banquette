package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/svenbs/banquette/pkg/handler"
)

var (
	addr     = flag.String("addr", ":8000", "sets the IP and Port to listen on.")
	dbaddr   = os.Getenv("DB_ADDR")
	dbuser   = os.Getenv("DB_USER")
	dbpass   = os.Getenv("DB_PASSWORD")
	dbsecret = os.Getenv("DB_SECRET")
	database = os.Getenv("DB_DATABASE")
)

func main() {
	flag.Parse()

	h, err := handler.InitDB("mysql", dbsecret, dbuser+":"+dbpass+"@("+dbaddr+")"+"/"+database)
	if err != nil {
		log.Fatalf("could not connect to database (@(%v)/%v): %v", dbaddr, database, err)
	}
	defer h.Close()

	loggedRouter := handlers.LoggingHandler(os.Stdout, serveHandler(h))

	log.Println("starting server on ", *addr)
	log.Fatal(http.ListenAndServe(*addr, loggedRouter))
	log.Println("Stopping...")
}

func serveHandler(env *handler.Env) http.Handler {

	r := mux.NewRouter()
	// sha256-token, username, [password]
	r.HandleFunc("/api/v1/oracle", env.OracleMethodRouter).Methods("POST", "DELETE")
	// dbtype, user, password, connectstring
	r.HandleFunc("/api/v1/token", env.TokenMethodRouter).Methods("POST", "PATCH", "DELETE")
	return r
}
