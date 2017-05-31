package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
	"time"
)

type GnucashXMLCommodity struct {
	Name        string `xml:"http://www.gnucash.org/XML/cmdty id"`
	Description string `xml:"http://www.gnucash.org/XML/cmdty name"`
	Type        string `xml:"http://www.gnucash.org/XML/cmdty space"`
	Fraction    int    `xml:"http://www.gnucash.org/XML/cmdty fraction"`
	XCode       string `xml:"http://www.gnucash.org/XML/cmdty xcode"`
}

type GnucashCommodity struct{ Security }

func (gc *GnucashCommodity) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var gxc GnucashXMLCommodity
	if err := d.DecodeElement(&gxc, &start); err != nil {
		return err
	}

	gc.Name = gxc.Name
	gc.Symbol = gxc.Name
	gc.Description = gxc.Description
	gc.AlternateId = gxc.XCode

	gc.Security.Type = Stock // assumed default
	if gxc.Type == "ISO4217" {
		gc.Security.Type = Currency
		// Get the number from our templates for the AlternateId because
		// Gnucash uses 'id' (our Name) to supply the string ISO4217 code
		template := FindSecurityTemplate(gxc.Name, Currency)
		if template == nil {
			return errors.New("Unable to find security template for Gnucash ISO4217 commodity")
		}
		gc.AlternateId = template.AlternateId
		gc.Precision = template.Precision
	} else {
		if gxc.Fraction > 0 {
			gc.Precision = int(math.Ceil(math.Log10(float64(gxc.Fraction))))
		} else {
			gc.Precision = 0
		}
	}
	return nil
}

type GnucashTime struct{ time.Time }

func (g *GnucashTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return fmt.Errorf("date should be a string")
	}
	t, err := time.Parse("2006-01-02 15:04:05 -0700", s)
	g.Time = t
	return err
}

type GnucashDate struct {
	Date GnucashTime `xml:"http://www.gnucash.org/XML/ts date"`
}

type GnucashAccount struct {
	Version         string              `xml:"version,attr"`
	accountid       int64               // Used to map Gnucash guid's to integer ones
	AccountId       string              `xml:"http://www.gnucash.org/XML/act id"`
	ParentAccountId string              `xml:"http://www.gnucash.org/XML/act parent"`
	Name            string              `xml:"http://www.gnucash.org/XML/act name"`
	Description     string              `xml:"http://www.gnucash.org/XML/act description"`
	Type            string              `xml:"http://www.gnucash.org/XML/act type"`
	Commodity       GnucashXMLCommodity `xml:"http://www.gnucash.org/XML/act commodity"`
}

type GnucashTransaction struct {
	TransactionId string              `xml:"http://www.gnucash.org/XML/trn id"`
	Description   string              `xml:"http://www.gnucash.org/XML/trn description"`
	DatePosted    GnucashDate         `xml:"http://www.gnucash.org/XML/trn date-posted"`
	DateEntered   GnucashDate         `xml:"http://www.gnucash.org/XML/trn date-entered"`
	Commodity     GnucashXMLCommodity `xml:"http://www.gnucash.org/XML/trn currency"`
	Splits        []GnucashSplit      `xml:"http://www.gnucash.org/XML/trn splits>split"`
}

type GnucashSplit struct {
	SplitId   string `xml:"http://www.gnucash.org/XML/split id"`
	Status    string `xml:"http://www.gnucash.org/XML/split reconciled-state"`
	AccountId string `xml:"http://www.gnucash.org/XML/split account"`
	Memo      string `xml:"http://www.gnucash.org/XML/split memo"`
	Amount    string `xml:"http://www.gnucash.org/XML/split quantity"`
	Value     string `xml:"http://www.gnucash.org/XML/split value"`
}

type GnucashXMLImport struct {
	XMLName      xml.Name             `xml:"gnc-v2"`
	Commodities  []GnucashCommodity   `xml:"http://www.gnucash.org/XML/gnc book>commodity"`
	Accounts     []GnucashAccount     `xml:"http://www.gnucash.org/XML/gnc book>account"`
	Transactions []GnucashTransaction `xml:"http://www.gnucash.org/XML/gnc book>transaction"`
}

type GnucashImport struct {
	Securities   []Security
	Accounts     []Account
	Transactions []Transaction
}

func ImportGnucash(r io.Reader) (*GnucashImport, error) {
	var gncxml GnucashXMLImport
	var gncimport GnucashImport

	// Perform initial parsing of xml into structs
	decoder := xml.NewDecoder(r)
	err := decoder.Decode(&gncxml)
	if err != nil {
		return nil, err
	}

	// Fixup securities, making a map of them as we go
	securityMap := make(map[string]Security)
	for i := range gncxml.Commodities {
		s := gncxml.Commodities[i].Security
		s.SecurityId = int64(i + 1)
		securityMap[s.Name] = s

		// Ignore gnucash's "template" commodity
		if s.Name != "template" ||
			s.Description != "template" ||
			s.AlternateId != "template" {
			gncimport.Securities = append(gncimport.Securities, s)
		}
	}

	//find root account, while simultaneously creating map of GUID's to
	//accounts
	var rootAccount GnucashAccount
	accountMap := make(map[string]GnucashAccount)
	for i := range gncxml.Accounts {
		gncxml.Accounts[i].accountid = int64(i + 1)
		if gncxml.Accounts[i].Type == "ROOT" {
			rootAccount = gncxml.Accounts[i]
		} else {
			accountMap[gncxml.Accounts[i].AccountId] = gncxml.Accounts[i]
		}
	}

	//Translate to our account format, figuring out parent relationships
	for guid := range accountMap {
		ga := accountMap[guid]
		var a Account

		a.AccountId = ga.accountid
		if ga.ParentAccountId == rootAccount.AccountId {
			a.ParentAccountId = -1
		} else {
			parent, ok := accountMap[ga.ParentAccountId]
			if ok {
				a.ParentAccountId = parent.accountid
			} else {
				a.ParentAccountId = -1 // Ugly, but assign to top-level if we can't find its parent
			}
		}
		a.Name = ga.Name
		if security, ok := securityMap[ga.Commodity.Name]; ok {
			a.SecurityId = security.SecurityId
		} else {
			return nil, fmt.Errorf("Unable to find security: %s", ga.Commodity.Name)
		}

		//TODO find account types
		switch ga.Type {
		default:
			a.Type = Bank
		case "ASSET":
			a.Type = Asset
		case "BANK":
			a.Type = Bank
		case "CASH":
			a.Type = Cash
		case "CREDIT", "LIABILITY":
			a.Type = Liability
		case "EQUITY":
			a.Type = Equity
		case "EXPENSE":
			a.Type = Expense
		case "INCOME":
			a.Type = Income
		case "PAYABLE":
			a.Type = Payable
		case "RECEIVABLE":
			a.Type = Receivable
		case "MUTUAL", "STOCK":
			a.Type = Investment
		case "TRADING":
			a.Type = Trading
		}

		gncimport.Accounts = append(gncimport.Accounts, a)
	}

	//Translate transactions to our format
	for i := range gncxml.Transactions {
		gt := gncxml.Transactions[i]

		t := new(Transaction)
		t.Description = gt.Description
		t.Date = gt.DatePosted.Date.Time
		for j := range gt.Splits {
			gs := gt.Splits[j]
			s := new(Split)
			s.Memo = gs.Memo

			switch gs.Status {
			default: // 'n', or not present
				s.Status = Imported
			case "c":
				s.Status = Cleared
			case "y":
				s.Status = Reconciled
			}

			account, ok := accountMap[gs.AccountId]
			if !ok {
				return nil, fmt.Errorf("Unable to find account: %s", gs.AccountId)
			}
			s.AccountId = account.accountid

			security, ok := securityMap[account.Commodity.Name]
			if !ok {
				return nil, fmt.Errorf("Unable to find security: %s", account.Commodity.Name)
			}
			s.SecurityId = -1

			var r big.Rat
			_, ok = r.SetString(gs.Amount)
			if ok {
				s.Amount = r.FloatString(security.Precision)
			} else {
				return nil, fmt.Errorf("Can't set split Amount: %s", gs.Amount)
			}

			t.Splits = append(t.Splits, s)
		}
		gncimport.Transactions = append(gncimport.Transactions, *t)
	}

	return &gncimport, nil
}

func GnucashImportHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		WriteError(w, 1 /*Not Signed In*/)
		return
	}

	if r.Method != "POST" {
		WriteError(w, 3 /*Invalid Request*/)
		return
	}

	multipartReader, err := r.MultipartReader()
	if err != nil {
		WriteError(w, 3 /*Invalid Request*/)
		return
	}

	// Assume there is only one 'part' and it's the one we care about
	part, err := multipartReader.NextPart()
	if err != nil {
		if err == io.EOF {
			WriteError(w, 3 /*Invalid Request*/)
		} else {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
		}
		return
	}

	gnucashImport, err := ImportGnucash(part)
	if err != nil {
		WriteError(w, 3 /*Invalid Request*/)
		return
	}

	sqltransaction, err := DB.Begin()
	if err != nil {
		WriteError(w, 999 /*Internal Error*/)
		log.Print(err)
		return
	}

	// Import securities, building map from Gnucash security IDs to our
	// internal IDs
	securityMap := make(map[int64]int64)
	for _, security := range gnucashImport.Securities {
		securityId := security.SecurityId // save off because it could be updated
		s, err := ImportGetCreateSecurity(sqltransaction, user, &security)
		if err != nil {
			sqltransaction.Rollback()
			WriteError(w, 6 /*Import Error*/)
			log.Print(err)
			log.Print(security)
			return
		}
		securityMap[securityId] = s.SecurityId
	}

	// Get/create accounts in the database, building a map from Gnucash account
	// IDs to our internal IDs as we go
	accountMap := make(map[int64]int64)
	accountsRemaining := len(gnucashImport.Accounts)
	accountsRemainingLast := accountsRemaining
	for accountsRemaining > 0 {
		for _, account := range gnucashImport.Accounts {

			// If the account has already been added to the map, skip it
			_, ok := accountMap[account.AccountId]
			if ok {
				continue
			}

			// If it hasn't been added, but its parent has, add it to the map
			_, ok = accountMap[account.ParentAccountId]
			if ok || account.ParentAccountId == -1 {
				account.UserId = user.UserId
				if account.ParentAccountId != -1 {
					account.ParentAccountId = accountMap[account.ParentAccountId]
				}
				account.SecurityId = securityMap[account.SecurityId]
				a, err := GetCreateAccountTx(sqltransaction, account)
				if err != nil {
					sqltransaction.Rollback()
					WriteError(w, 999 /*Internal Error*/)
					log.Print(err)
					return
				}
				accountMap[account.AccountId] = a.AccountId
				accountsRemaining--
			}
		}
		if accountsRemaining == accountsRemainingLast {
			//We didn't make any progress in importing the next level of accounts, so there must be a circular parent-child relationship, so give up and tell the user they're wrong
			sqltransaction.Rollback()
			WriteError(w, 999 /*Internal Error*/)
			log.Print(fmt.Errorf("Circular account parent-child relationship when importing %s", part.FileName()))
			return
		}
		accountsRemainingLast = accountsRemaining
	}

	// Insert transactions, fixing up account IDs to match internal ones from
	// above
	for _, transaction := range gnucashImport.Transactions {
		for _, split := range transaction.Splits {
			acctId, ok := accountMap[split.AccountId]
			if !ok {
				sqltransaction.Rollback()
				WriteError(w, 999 /*Internal Error*/)
				log.Print(fmt.Errorf("Error: Split's AccountID Doesn't exist: %d\n", split.AccountId))
				return
			}
			split.AccountId = acctId
		}
		err := InsertTransactionTx(sqltransaction, &transaction, user)
		if err != nil {
			sqltransaction.Rollback()
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
	}

	err = sqltransaction.Commit()
	if err != nil {
		sqltransaction.Rollback()
		WriteError(w, 999 /*Internal Error*/)
		log.Print(err)
		return
	}

	WriteSuccess(w)
}
