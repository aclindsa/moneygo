package handlers

import (
	"gopkg.in/gorp.v1"
	"log"
	"net/http"
	"path"
	"strings"
)

// But who writes the ResponseWriterWriter?
type ResponseWriterWriter interface {
	Write(http.ResponseWriter) error
}
type Tx = gorp.Transaction
type Context struct {
	Tx        *Tx
	User      *User
	Remaining string // portion of URL not yet reached in the hierarchy
}
type Handler func(*http.Request, *Context) ResponseWriterWriter

func NextLevel(previous string) (current, remaining string) {
	split := strings.SplitN(previous, "/", 2)
	if len(split) == 2 {
		return split[0], split[1]
	}
	return split[0], ""
}

type APIHandler struct {
	DB *gorp.DbMap
}

func (ah *APIHandler) txWrapper(h Handler, r *http.Request, context *Context) (writer ResponseWriterWriter) {
	tx, err := ah.DB.Begin()
	if err != nil {
		log.Print(err)
		return NewError(999 /*Internal Error*/)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if _, ok := writer.(*Error); ok {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				log.Print(err)
				writer = NewError(999 /*Internal Error*/)
			}
		}
	}()

	context.Tx = tx
	return h(r, context)
}

func (ah *APIHandler) route(r *http.Request) ResponseWriterWriter {
	current, remaining := NextLevel(path.Clean("/" + r.URL.Path)[1:])
	if current != "v1" {
		return NewError(3 /*Invalid Request*/)
	}

	current, remaining = NextLevel(remaining)
	context := &Context{Remaining: remaining}

	switch current {
	case "sessions":
		return ah.txWrapper(SessionHandler, r, context)
	case "users":
		return ah.txWrapper(UserHandler, r, context)
	case "securities":
		return ah.txWrapper(SecurityHandler, r, context)
	case "securitytemplates":
		return SecurityTemplateHandler(r, context)
	case "prices":
		return ah.txWrapper(PriceHandler, r, context)
	case "accounts":
		return ah.txWrapper(AccountHandler, r, context)
	case "transactions":
		return ah.txWrapper(TransactionHandler, r, context)
	case "imports":
		return ah.txWrapper(ImportHandler, r, context)
	case "reports":
		return ah.txWrapper(ReportHandler, r, context)
	default:
		return NewError(3 /*Invalid Request*/)
	}
}

func (ah *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.route(r).Write(w)
}
