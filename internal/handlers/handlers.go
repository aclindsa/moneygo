package handlers

import (
	"gopkg.in/gorp.v1"
	"log"
	"net/http"
)

// But who writes the ResponseWriterWriter?
type ResponseWriterWriter interface {
	Write(http.ResponseWriter) error
}
type Tx = gorp.Transaction
type TxHandler func(*http.Request, *Tx) ResponseWriterWriter

func TxHandlerFunc(t TxHandler, db *gorp.DbMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tx, err := db.Begin()
		if err != nil {
			log.Print(err)
			WriteError(w, 999 /*Internal Error*/)
			return
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				WriteError(w, 999 /*Internal Error*/)
				panic(r)
			}
		}()

		writer := t(r, tx)

		if e, ok := writer.(*Error); ok {
			tx.Rollback()
			e.Write(w)
		} else {
			err = tx.Commit()
			if err != nil {
				log.Print(err)
				WriteError(w, 999 /*Internal Error*/)
			} else {
				err = writer.Write(w)
				if err != nil {
					log.Print(err)
					WriteError(w, 999 /*Internal Error*/)
				}
			}
		}
	}
}

func GetHandler(db *gorp.DbMap) *http.ServeMux {
	servemux := http.NewServeMux()
	servemux.HandleFunc("/v1/sessions/", TxHandlerFunc(SessionHandler, db))
	servemux.HandleFunc("/v1/users/", TxHandlerFunc(UserHandler, db))
	servemux.HandleFunc("/v1/securities/", TxHandlerFunc(SecurityHandler, db))
	servemux.HandleFunc("/v1/prices/", TxHandlerFunc(PriceHandler, db))
	servemux.HandleFunc("/v1/securitytemplates/", SecurityTemplateHandler)
	servemux.HandleFunc("/v1/accounts/", TxHandlerFunc(AccountHandler, db))
	servemux.HandleFunc("/v1/transactions/", TxHandlerFunc(TransactionHandler, db))
	servemux.HandleFunc("/v1/imports/gnucash", TxHandlerFunc(GnucashImportHandler, db))
	servemux.HandleFunc("/v1/reports/", TxHandlerFunc(ReportHandler, db))

	return servemux
}
