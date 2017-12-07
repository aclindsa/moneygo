package handlers

//go:generate make

import (
	"errors"
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store/db"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func SearchSecurityTemplates(search string, _type models.SecurityType, limit int64) []*models.Security {
	upperSearch := strings.ToUpper(search)
	var results []*models.Security
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

func FindSecurityTemplate(name string, _type models.SecurityType) *models.Security {
	for _, security := range SecurityTemplates {
		if name == security.Name && _type == security.Type {
			return &security
		}
	}
	return nil
}

func FindCurrencyTemplate(iso4217 int64) *models.Security {
	iso4217string := strconv.FormatInt(iso4217, 10)
	for _, security := range SecurityTemplates {
		if security.Type == models.Currency && security.AlternateId == iso4217string {
			return &security
		}
	}
	return nil
}

func GetSecurity(tx *db.Tx, securityid int64, userid int64) (*models.Security, error) {
	var s models.Security

	err := tx.SelectOne(&s, "SELECT * from securities where UserId=? AND SecurityId=?", userid, securityid)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func GetSecurities(tx *db.Tx, userid int64) (*[]*models.Security, error) {
	var securities []*models.Security

	_, err := tx.Select(&securities, "SELECT * from securities where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &securities, nil
}

func InsertSecurity(tx *db.Tx, s *models.Security) error {
	err := tx.Insert(s)
	if err != nil {
		return err
	}
	return nil
}

func UpdateSecurity(tx *db.Tx, s *models.Security) (err error) {
	user, err := GetUser(tx, s.UserId)
	if err != nil {
		return
	} else if user.DefaultCurrency == s.SecurityId && s.Type != models.Currency {
		return errors.New("Cannot change security which is user's default currency to be non-currency")
	}

	count, err := tx.Update(s)
	if err != nil {
		return
	}
	if count > 1 {
		return fmt.Errorf("Updated %d securities (expected 1)", count)
	}

	return nil
}

type SecurityInUseError struct {
	message string
}

func (e SecurityInUseError) Error() string {
	return e.message
}

func DeleteSecurity(tx *db.Tx, s *models.Security) error {
	// First, ensure no accounts are using this security
	accounts, err := tx.SelectInt("SELECT count(*) from accounts where UserId=? and SecurityId=?", s.UserId, s.SecurityId)

	if accounts != 0 {
		return SecurityInUseError{"One or more accounts still use this security"}
	}

	user, err := GetUser(tx, s.UserId)
	if err != nil {
		return err
	} else if user.DefaultCurrency == s.SecurityId {
		return SecurityInUseError{"Cannot delete security which is user's default currency"}
	}

	// Remove all prices involving this security (either of this security, or
	// using it as a currency)
	_, err = tx.Exec("DELETE FROM prices WHERE SecurityId=? OR CurrencyId=?", s.SecurityId, s.SecurityId)
	if err != nil {
		return err
	}

	count, err := tx.Delete(s)
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Deleted more than one security")
	}

	return nil
}

func ImportGetCreateSecurity(tx *db.Tx, userid int64, security *models.Security) (*models.Security, error) {
	security.UserId = userid
	if len(security.AlternateId) == 0 {
		// Always create a new local security if we can't match on the AlternateId
		err := InsertSecurity(tx, security)
		if err != nil {
			return nil, err
		}
		return security, nil
	}

	var securities []*models.Security

	_, err := tx.Select(&securities, "SELECT * from securities where UserId=? AND Type=? AND AlternateId=? AND Preciseness=?", userid, security.Type, security.AlternateId, security.Precision)
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
	err = InsertSecurity(tx, security)
	if err != nil {
		return nil, err
	}

	return security, nil
}

func SecurityHandler(r *http.Request, context *Context) ResponseWriterWriter {
	user, err := GetUserFromSession(context.Tx, r)
	if err != nil {
		return NewError(1 /*Not Signed In*/)
	}

	if r.Method == "POST" {
		if !context.LastLevel() {
			securityid, err := context.NextID()
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}
			if context.NextLevel() != "prices" {
				return NewError(3 /*Invalid Request*/)
			}
			return PriceHandler(r, context, user, securityid)
		}

		var security models.Security
		if err := ReadJSON(r, &security); err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		security.SecurityId = -1
		security.UserId = user.UserId

		err = InsertSecurity(context.Tx, &security)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}

		return ResponseWrapper{201, &security}
	} else if r.Method == "GET" {
		if context.LastLevel() {
			//Return all securities
			var sl models.SecurityList

			securities, err := GetSecurities(context.Tx, user.UserId)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			sl.Securities = securities
			return &sl
		} else {
			securityid, err := context.NextID()
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			if !context.LastLevel() {
				if context.NextLevel() != "prices" {
					return NewError(3 /*Invalid Request*/)
				}
				return PriceHandler(r, context, user, securityid)
			}

			security, err := GetSecurity(context.Tx, securityid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			return security
		}
	} else {
		securityid, err := context.NextID()
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		if !context.LastLevel() {
			if context.NextLevel() != "prices" {
				return NewError(3 /*Invalid Request*/)
			}
			return PriceHandler(r, context, user, securityid)
		}

		if r.Method == "PUT" {
			var security models.Security
			if err := ReadJSON(r, &security); err != nil || security.SecurityId != securityid {
				return NewError(3 /*Invalid Request*/)
			}
			security.UserId = user.UserId

			err = UpdateSecurity(context.Tx, &security)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return &security
		} else if r.Method == "DELETE" {
			security, err := GetSecurity(context.Tx, securityid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			err = DeleteSecurity(context.Tx, security)
			if _, ok := err.(SecurityInUseError); ok {
				return NewError(7 /*In Use Error*/)
			} else if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}

func SecurityTemplateHandler(r *http.Request, context *Context) ResponseWriterWriter {
	if r.Method == "GET" {
		var sl models.SecurityList

		query, _ := url.ParseQuery(r.URL.RawQuery)

		var limit int64 = -1
		search := query.Get("search")

		var _type models.SecurityType = 0
		typestring := query.Get("type")
		if len(typestring) > 0 {
			_type = models.GetSecurityType(typestring)
			if _type == 0 {
				return NewError(3 /*Invalid Request*/)
			}
		}

		limitstring := query.Get("limit")
		if limitstring != "" {
			limitint, err := strconv.ParseInt(limitstring, 10, 0)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}
			limit = limitint
		}

		securities := SearchSecurityTemplates(search, _type, limit)

		sl.Securities = &securities
		return &sl
	} else {
		return NewError(3 /*Invalid Request*/)
	}
}
