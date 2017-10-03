package main

//go:generate make

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path"
	"strconv"
)

var configFile string
var config *Config

func init() {
	var err error
	flag.StringVar(&configFile, "config", "/etc/moneygo/config.ini", "Path to config file")
	flag.Parse()

	config, err = readConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	static_path := path.Join(config.MoneyGo.Basedir, "static")

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

	initDB(config)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join(config.MoneyGo.Basedir, "static/index.html"))
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join(config.MoneyGo.Basedir, r.URL.Path))
}

func main() {
	servemux := http.NewServeMux()
	servemux.HandleFunc("/", rootHandler)
	servemux.HandleFunc("/static/", staticHandler)
	servemux.HandleFunc("/session/", SessionHandler)
	servemux.HandleFunc("/user/", UserHandler)
	servemux.HandleFunc("/security/", SecurityHandler)
	servemux.HandleFunc("/securitytemplate/", SecurityTemplateHandler)
	servemux.HandleFunc("/account/", AccountHandler)
	servemux.HandleFunc("/transaction/", TransactionHandler)
	servemux.HandleFunc("/import/gnucash", GnucashImportHandler)
	servemux.HandleFunc("/report/", ReportHandler)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(config.MoneyGo.Port))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Serving on port %d out of directory: %s", config.MoneyGo.Port, config.MoneyGo.Basedir)
	if config.MoneyGo.Fcgi {
		fcgi.Serve(listener, servemux)
	} else {
		http.Serve(listener, servemux)
	}
}
