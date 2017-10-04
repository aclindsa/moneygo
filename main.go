package main

//go:generate make

import (
	"database/sql"
	"flag"
	"github.com/aclindsa/moneygo/internal/config"
	"github.com/aclindsa/moneygo/internal/db"
	"github.com/aclindsa/moneygo/internal/handlers"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path"
	"strconv"
)

var configFile string
var cfg *config.Config

func init() {
	var err error
	flag.StringVar(&configFile, "config", "/etc/moneygo/config.ini", "Path to config file")
	flag.Parse()

	cfg, err = config.ReadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	static_path := path.Join(cfg.MoneyGo.Basedir, "static")

	// Ensure base directory is valid
	dir_err_str := "The base directory doesn't look like it contains the " +
		"'static' directory. Check to make sure your config file contains the" +
		"right path for 'base-directory'."
	static_dir, err := os.Stat(static_path)
	if err != nil {
		log.Print(err)
		log.Fatal(dir_err_str)
	}
	if !static_dir.IsDir() {
		log.Fatal(dir_err_str)
	}

	// Setup the logging flags to be printed
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

type FileHandler func(http.ResponseWriter, *http.Request, string)

func FileHandlerFunc(h FileHandler, basedir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r, basedir)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request, basedir string) {
	http.ServeFile(w, r, path.Join(basedir, "static/index.html"))
}

func staticHandler(w http.ResponseWriter, r *http.Request, basedir string) {
	http.ServeFile(w, r, path.Join(basedir, r.URL.Path))
}

func main() {
	database, err := sql.Open(cfg.MoneyGo.DBType.String(), cfg.MoneyGo.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	dbmap, err := db.GetDbMap(database, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Get ServeMux for API and add our own handlers for files
	servemux := handlers.GetHandler(dbmap)
	servemux.HandleFunc("/", FileHandlerFunc(rootHandler, cfg.MoneyGo.Basedir))
	servemux.HandleFunc("/static/", FileHandlerFunc(staticHandler, cfg.MoneyGo.Basedir))

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.MoneyGo.Port))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Serving on port %d out of directory: %s", cfg.MoneyGo.Port, cfg.MoneyGo.Basedir)
	if cfg.MoneyGo.Fcgi {
		fcgi.Serve(listener, servemux)
	} else {
		http.Serve(listener, servemux)
	}
}
