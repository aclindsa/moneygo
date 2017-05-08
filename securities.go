package main

import (
	"encoding/json"
	"errors"
	"gopkg.in/gorp.v1"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	Currency int64 = 1
	Stock          = 2
)

func GetSecurityType(typestring string) int64 {
	if strings.EqualFold(typestring, "currency") {
		return Currency
	} else if strings.EqualFold(typestring, "stock") {
		return Stock
	} else {
		return 0
	}
}

type Security struct {
	SecurityId  int64
	UserId      int64
	Name        string
	Description string
	Symbol      string
	// Number of decimal digits (to the right of the decimal point) this
	// security is precise to
	Precision int
	Type      int64
	// AlternateId is CUSIP for Type=Stock
	AlternateId string
}

type SecurityList struct {
	Securities *[]*Security `json:"securities"`
}

func (s *Security) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(s)
}

func (s *Security) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s)
}

func (sl *SecurityList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(sl)
}

func SearchSecurityTemplates(search string, _type int64, limit int64) []*Security {
	upperSearch := strings.ToUpper(search)
	var results []*Security
	for i, security := range SecurityTemplates {
		if strings.Contains(strings.ToUpper(security.Name), upperSearch) ||
			strings.Contains(strings.ToUpper(security.Description), upperSearch) ||
			strings.Contains(strings.ToUpper(security.Symbol), upperSearch) {
			if _type == 0 || _type == security.Type {
				results = append(results, &SecurityTemplates[i])
				if limit != -1 && int64(len(results)) >= limit {
					break
				}
			}
		}
	}
	return results
}

func FindSecurityTemplate(name string, _type int64) *Security {
	for _, security := range SecurityTemplates {
		if name == security.Name && _type == security.Type {
			return &security
		}
	}
	return nil
}

func GetSecurity(securityid int64, userid int64) (*Security, error) {
	var s Security

	err := DB.SelectOne(&s, "SELECT * from securities where UserId=? AND SecurityId=?", userid, securityid)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func GetSecurities(userid int64) (*[]*Security, error) {
	var securities []*Security

	_, err := DB.Select(&securities, "SELECT * from securities where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &securities, nil
}

func InsertSecurity(s *Security) error {
	err := DB.Insert(s)
	if err != nil {
		return err
	}
	return nil
}

func InsertSecurityTx(transaction *gorp.Transaction, s *Security) error {
	err := transaction.Insert(s)
	if err != nil {
		return err
	}
	return nil
}

func UpdateSecurity(s *Security) error {
	transaction, err := DB.Begin()
	if err != nil {
		return err
	}

	count, err := transaction.Update(s)
	if err != nil {
		transaction.Rollback()
		return err
	}
	if count != 1 {
		transaction.Rollback()
		return errors.New("Updated more than one security")
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return err
	}

	return nil
}

func DeleteSecurity(s *Security) error {
	transaction, err := DB.Begin()
	if err != nil {
		return err
	}

	// First, ensure no accounts are using this security
	accounts, err := transaction.SelectInt("SELECT count(*) from accounts where UserId=? and SecurityId=?", s.UserId, s.SecurityId)

	if accounts != 0 {
		transaction.Rollback()
		return errors.New("One or more accounts still use this security")
	}

	count, err := transaction.Delete(s)
	if err != nil {
		transaction.Rollback()
		return err
	}
	if count != 1 {
		transaction.Rollback()
		return errors.New("Deleted more than one security")
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return err
	}

	return nil
}

func ImportGetCreateSecurity(transaction *gorp.Transaction, user *User, security *Security) (*Security, error) {
	security.UserId = user.UserId
	if len(security.AlternateId) == 0 {
		// Always create a new local security if we can't match on the AlternateId
		err := InsertSecurityTx(transaction, security)
		if err != nil {
			return nil, err
		}
		return security, nil
	}

	var securities []*Security

	_, err := transaction.Select(&securities, "SELECT * from securities where UserId=? AND Type=? AND AlternateId=? AND Precision=?", user.UserId, security.Type, security.AlternateId, security.Precision)
	if err != nil {
		return nil, err
	}

	// First try to find a case insensitive match on the name or symbol
	upperName := strings.ToUpper(security.Name)
	upperSymbol := strings.ToUpper(security.Symbol)
	for _, s := range securities {
		if (len(s.Name) > 0 && strings.ToUpper(s.Name) == upperName) ||
			(len(s.Symbol) > 0 && strings.ToUpper(s.Symbol) == upperSymbol) {
			return s, nil
		}
	}
	//		if strings.Contains(strings.ToUpper(security.Name), upperSearch) ||

	// Try to find a partial string match on the name or symbol
	for _, s := range securities {
		sUpperName := strings.ToUpper(s.Name)
		sUpperSymbol := strings.ToUpper(s.Symbol)
		if (len(upperName) > 0 && len(s.Name) > 0 && (strings.Contains(upperName, sUpperName) || strings.Contains(sUpperName, upperName))) ||
			(len(upperSymbol) > 0 && len(s.Symbol) > 0 && (strings.Contains(upperSymbol, sUpperSymbol) || strings.Contains(sUpperSymbol, upperSymbol))) {
			return s, nil
		}
	}

	// Give up and return the first security in the list
	if len(securities) > 0 {
		return securities[0], nil
	}

	// If there wasn't even one security in the list, make a new one
	err = InsertSecurityTx(transaction, security)
	if err != nil {
		return nil, err
	}

	return security, nil
}

func SecurityHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		WriteError(w, 1 /*Not Signed In*/)
		return
	}

	if r.Method == "POST" {
		security_json := r.PostFormValue("security")
		if security_json == "" {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		var security Security
		err := security.Read(security_json)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		security.SecurityId = -1
		security.UserId = user.UserId

		err = InsertSecurity(&security)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}

		w.WriteHeader(201 /*Created*/)
		err = security.Write(w)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
	} else if r.Method == "GET" {
		var securityid int64
		n, err := GetURLPieces(r.URL.Path, "/security/%d", &securityid)

		if err != nil || n != 1 {
			//Return all securities
			var sl SecurityList

			securities, err := GetSecurities(user.UserId)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			sl.Securities = securities
			err = (&sl).Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		} else {
			security, err := GetSecurity(securityid, user.UserId)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			err = security.Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		}
	} else {
		securityid, err := GetURLID(r.URL.Path)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		if r.Method == "PUT" {
			security_json := r.PostFormValue("security")
			if security_json == "" {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			var security Security
			err := security.Read(security_json)
			if err != nil || security.SecurityId != securityid {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}
			security.UserId = user.UserId

			err = UpdateSecurity(&security)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			err = security.Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		} else if r.Method == "DELETE" {
			security, err := GetSecurity(securityid, user.UserId)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			err = DeleteSecurity(security)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			WriteSuccess(w)
		}
	}
}

func SecurityTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var sl SecurityList

		query, _ := url.ParseQuery(r.URL.RawQuery)

		var limit int64 = -1
		search := query.Get("search")
		_type := GetSecurityType(query.Get("type"))

		limitstring := query.Get("limit")
		if limitstring != "" {
			limitint, err := strconv.ParseInt(limitstring, 10, 0)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}
			limit = limitint
		}

		securities := SearchSecurityTemplates(search, _type, limit)

		sl.Securities = &securities
		err := (&sl).Write(w)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
	} else {
		WriteError(w, 3 /*Invalid Request*/)
	}
}
