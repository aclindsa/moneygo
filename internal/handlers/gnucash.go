package handlers

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"io"
	"log"
	"math"
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

type GnucashCommodity struct{ models.Security }

func (gc *GnucashCommodity) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var gxc GnucashXMLCommodity
	if err := d.DecodeElement(&gxc, &start); err != nil {
		return err
	}

	gc.Name = gxc.Name
	gc.Symbol = gxc.Name
	gc.Description = gxc.Description
	gc.AlternateId = gxc.XCode

	gc.Security.Type = models.Stock // assumed default
	if gxc.Type == "ISO4217" || gxc.Type == "CURRENCY" {
		gc.Security.Type = models.Currency
		// Get the number from our templates for the AlternateId because
		// Gnucash uses 'id' (our Name) to supply the string ISO4217 code
		template := FindSecurityTemplate(gxc.Name, models.Currency)
		if template == nil {
			return errors.New("Unable to find security template for Gnucash ISO4217 commodity")
		}
		gc.AlternateId = template.AlternateId
		gc.Precision = template.Precision
	} else {
		if gxc.Fraction > 0 {
			gc.Precision = uint64(math.Ceil(math.Log10(float64(gxc.Fraction))))
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

type GnucashPrice struct {
	Id        string           `xml:"http://www.gnucash.org/XML/price id"`
	Commodity GnucashCommodity `xml:"http://www.gnucash.org/XML/price commodity"`
	Currency  GnucashCommodity `xml:"http://www.gnucash.org/XML/price currency"`
	Date      GnucashDate      `xml:"http://www.gnucash.org/XML/price time"`
	Source    string           `xml:"http://www.gnucash.org/XML/price source"`
	Type      string           `xml:"http://www.gnucash.org/XML/price type"`
	Value     string           `xml:"http://www.gnucash.org/XML/price value"`
}

type GnucashPriceDB struct {
	Prices []GnucashPrice `xml:"price"`
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
	Number        string              `xml:"http://www.gnucash.org/XML/trn num"`
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
	PriceDB      GnucashPriceDB       `xml:"http://www.gnucash.org/XML/gnc book>pricedb"`
	Accounts     []GnucashAccount     `xml:"http://www.gnucash.org/XML/gnc book>account"`
	Transactions []GnucashTransaction `xml:"http://www.gnucash.org/XML/gnc book>transaction"`
}

type GnucashImport struct {
	Securities   []models.Security
	Accounts     []models.Account
	Transactions []models.Transaction
	Prices       []models.Price
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
	securityMap := make(map[string]models.Security)
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

	// Create prices, setting security and currency IDs from securityMap
	for i := range gncxml.PriceDB.Prices {
		price := gncxml.PriceDB.Prices[i]
		var p models.Price
		security, ok := securityMap[price.Commodity.Name]
		if !ok {
			return nil, fmt.Errorf("Unable to find commodity '%s' for price '%s'", price.Commodity.Name, price.Id)
		}
		currency, ok := securityMap[price.Currency.Name]
		if !ok {
			return nil, fmt.Errorf("Unable to find currency '%s' for price '%s'", price.Currency.Name, price.Id)
		}
		if currency.Type != models.Currency {
			return nil, fmt.Errorf("Currency for imported price isn't actually a currency\n")
		}
		p.PriceId = int64(i + 1)
		p.SecurityId = security.SecurityId
		p.CurrencyId = currency.SecurityId
		p.Date = price.Date.Date.Time

		_, ok = p.Value.SetString(price.Value)
		if !ok {
			return nil, fmt.Errorf("Can't set price value: %s", price.Value)
		}
		if p.Value.Precision() > currency.Precision {
			// TODO we're possibly losing data here... but do we care?
			p.Value.Round(currency.Precision)
		}

		p.RemoteId = "gnucash:" + price.Id
		gncimport.Prices = append(gncimport.Prices, p)
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
		var a models.Account

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
			a.Type = models.Bank
		case "ASSET":
			a.Type = models.Asset
		case "BANK":
			a.Type = models.Bank
		case "CASH":
			a.Type = models.Cash
		case "CREDIT", "LIABILITY":
			a.Type = models.Liability
		case "EQUITY":
			a.Type = models.Equity
		case "EXPENSE":
			a.Type = models.Expense
		case "INCOME":
			a.Type = models.Income
		case "PAYABLE":
			a.Type = models.Payable
		case "RECEIVABLE":
			a.Type = models.Receivable
		case "MUTUAL", "STOCK":
			a.Type = models.Investment
		case "TRADING":
			a.Type = models.Trading
		}

		gncimport.Accounts = append(gncimport.Accounts, a)
	}

	//Translate transactions to our format
	for i := range gncxml.Transactions {
		gt := gncxml.Transactions[i]

		t := new(models.Transaction)
		t.Description = gt.Description
		t.Date = gt.DatePosted.Date.Time
		for j := range gt.Splits {
			gs := gt.Splits[j]
			s := new(models.Split)

			switch gs.Status {
			default: // 'n', or not present
				s.Status = models.Imported
			case "c":
				s.Status = models.Cleared
			case "y":
				s.Status = models.Reconciled
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

			s.RemoteId = "gnucash:" + gs.SplitId
			s.Number = gt.Number
			s.Memo = gs.Memo

			_, ok = s.Amount.SetString(gs.Amount)
			if !ok {
				return nil, fmt.Errorf("Can't set split Amount: %s", gs.Amount)
			}
			if s.Amount.Precision() > security.Precision {
				return nil, fmt.Errorf("Imported price's precision (%d) is greater than the security's (%s)\n", s.Amount.Precision(), security)
			}

			t.Splits = append(t.Splits, s)
		}
		gncimport.Transactions = append(gncimport.Transactions, *t)
	}

	return &gncimport, nil
}

func GnucashImportHandler(r *http.Request, context *Context) ResponseWriterWriter {
	user, err := GetUserFromSession(context.Tx, r)
	if err != nil {
		return NewError(1 /*Not Signed In*/)
	}

	if r.Method != "POST" {
		return NewError(3 /*Invalid Request*/)
	}

	multipartReader, err := r.MultipartReader()
	if err != nil {
		return NewError(3 /*Invalid Request*/)
	}

	// Assume there is only one 'part' and it's the one we care about
	part, err := multipartReader.NextPart()
	if err != nil {
		if err == io.EOF {
			return NewError(3 /*Invalid Request*/)
		} else {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}
	}

	bufread := bufio.NewReader(part)
	gzHeader, err := bufread.Peek(2)
	if err != nil {
		log.Print(err)
		return NewError(999 /*Internal Error*/)
	}

	// Does this look like a gzipped file?
	var gnucashImport *GnucashImport
	if gzHeader[0] == 0x1f && gzHeader[1] == 0x8b {
		gzr, err2 := gzip.NewReader(bufread)
		if err2 != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}
		gnucashImport, err = ImportGnucash(gzr)
	} else {
		gnucashImport, err = ImportGnucash(bufread)
	}

	if err != nil {
		log.Print(err)
		return NewError(3 /*Invalid Request*/)
	}

	// Import securities, building map from Gnucash security IDs to our
	// internal IDs
	securityMap := make(map[int64]int64)
	for _, security := range gnucashImport.Securities {
		securityId := security.SecurityId // save off because it could be updated
		s, err := ImportGetCreateSecurity(context.Tx, user.UserId, &security)
		if err != nil {
			log.Print(err)
			log.Print(security)
			return NewError(6 /*Import Error*/)
		}
		securityMap[securityId] = s.SecurityId
	}

	// Import prices, setting security and currency IDs from securityMap
	for _, price := range gnucashImport.Prices {
		price.SecurityId = securityMap[price.SecurityId]
		price.CurrencyId = securityMap[price.CurrencyId]
		price.PriceId = 0

		err := CreatePriceIfNotExist(context.Tx, &price)
		if err != nil {
			log.Print(err)
			return NewError(6 /*Import Error*/)
		}
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
				a, err := GetCreateAccount(context.Tx, account)
				if err != nil {
					log.Print(err)
					return NewError(999 /*Internal Error*/)
				}
				accountMap[account.AccountId] = a.AccountId
				accountsRemaining--
			}
		}
		if accountsRemaining == accountsRemainingLast {
			//We didn't make any progress in importing the next level of accounts, so there must be a circular parent-child relationship, so give up and tell the user they're wrong
			log.Print(fmt.Errorf("Circular account parent-child relationship when importing %s", part.FileName()))
			return NewError(999 /*Internal Error*/)
		}
		accountsRemainingLast = accountsRemaining
	}

	// Insert transactions, fixing up account IDs to match internal ones from
	// above
	for _, transaction := range gnucashImport.Transactions {
		var already_imported bool
		for _, split := range transaction.Splits {
			acctId, ok := accountMap[split.AccountId]
			if !ok {
				log.Print(fmt.Errorf("Error: Split's AccountID Doesn't exist: %d\n", split.AccountId))
				return NewError(999 /*Internal Error*/)
			}
			split.AccountId = acctId

			exists, err := context.Tx.SplitExists(split)
			if err != nil {
				log.Print("Error checking if split was already imported:", err)
				return NewError(999 /*Internal Error*/)
			} else if exists {
				already_imported = true
			}
		}
		if !already_imported {
			err := context.Tx.InsertTransaction(&transaction, user)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
		}
	}

	return SuccessWriter{}
}
