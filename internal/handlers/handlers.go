package handlers

import (
	"gopkg.in/gorp.v1"
	"net/http"
)

// Create a closure over db, allowing the handlers to look like a
// http.HandlerFunc
type DB = gorp.DbMap
type DBHandler func(http.ResponseWriter, *http.Request, *DB)

func DBHandlerFunc(h DBHandler, db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r, db)
	}
}

func GetHandler(db *DB) *http.ServeMux {
	servemux := http.NewServeMux()
	servemux.HandleFunc("/session/", DBHandlerFunc(SessionHandler, db))
	servemux.HandleFunc("/user/", DBHandlerFunc(UserHandler, db))
	servemux.HandleFunc("/security/", DBHandlerFunc(SecurityHandler, db))
	servemux.HandleFunc("/securitytemplate/", SecurityTemplateHandler)
	servemux.HandleFunc("/account/", DBHandlerFunc(AccountHandler, db))
	servemux.HandleFunc("/transaction/", DBHandlerFunc(TransactionHandler, db))
	servemux.HandleFunc("/import/gnucash", DBHandlerFunc(GnucashImportHandler, db))
	servemux.HandleFunc("/report/", DBHandlerFunc(ReportHandler, db))

	return servemux
}
