package main

import (
	"errors"
	"fmt"
	"github.com/aclindsa/ofxgo"
	"io"
	"math/big"
)

type OFXImport struct {
	Securities   []Security
	Accounts     []Account
	Transactions []Transaction
	//	Balances     map[int64]string // map AccountIDs to ending balances
}

func (i *OFXImport) GetSecurity(ofxsecurityid int64) (*Security, error) {
	if ofxsecurityid < 0 || ofxsecurityid > int64(len(i.Securities)) {
		return nil, errors.New("OFXImport.GetSecurity: SecurityID out of range")
	}
	return &i.Securities[ofxsecurityid], nil
}

func (i *OFXImport) GetAddCurrency(isoname string) (*Security, error) {
	for _, security := range i.Securities {
		if isoname == security.Name && Currency == security.Type {
			return &security, nil
		}
	}

	template := FindSecurityTemplate(isoname, Currency)
	if template == nil {
		return nil, fmt.Errorf("Failed to find Security for \"%s\"", isoname)
	}
	var security Security = *template
	security.SecurityId = int64(len(i.Securities) + 1)
	i.Securities = append(i.Securities, security)

	return &security, nil
}

func (i *OFXImport) AddTransaction(tran *ofxgo.Transaction, account *Account) error {
	var t Transaction

	t.Date = tran.DtPosted.UTC()
	t.RemoteId = tran.FiTID.String()
	// TODO CorrectFiTID/CorrectAction?
	// Construct the description from whichever of the descriptive OFX fields are present
	if len(tran.Name) > 0 {
		t.Description = string(tran.Name)
	} else if tran.Payee != nil {
		t.Description = string(tran.Payee.Name)
	}
	if len(tran.Memo) > 0 {
		if len(t.Description) > 0 {
			t.Description = t.Description + " - " + string(tran.Memo)
		} else {
			t.Description = string(tran.Memo)
		}
	}

	var s1, s2 Split
	if len(tran.ExtdName) > 0 {
		s1.Memo = tran.ExtdName.String()
	}
	if len(tran.CheckNum) > 0 {
		s1.Number = tran.CheckNum.String()
	} else if len(tran.RefNum) > 0 {
		s1.Number = tran.RefNum.String()
	}

	amt := big.NewRat(0, 1)
	// Convert TrnAmt to account's currency if Currency is set
	if ok, _ := tran.Currency.Valid(); ok {
		amt.Mul(&tran.Currency.CurRate.Rat, &tran.TrnAmt.Rat)
	} else {
		amt.Set(&tran.TrnAmt.Rat)
	}
	if account.SecurityId < 1 || account.SecurityId > int64(len(i.Securities)) {
		return errors.New("Internal error: security index not found in OFX import\n")
	}
	security := i.Securities[account.SecurityId-1]
	s1.Amount = amt.FloatString(security.Precision)
	s2.Amount = amt.Neg(amt).FloatString(security.Precision)

	s1.Status = Imported
	s2.Status = Imported

	s1.AccountId = account.AccountId
	s2.AccountId = -1
	s1.SecurityId = -1
	s2.SecurityId = security.SecurityId

	t.Splits = append(t.Splits, &s1)
	t.Splits = append(t.Splits, &s2)
	i.Transactions = append(i.Transactions, t)

	return nil
}

func (i *OFXImport) importOFXBank(stmt *ofxgo.StatementResponse) error {
	security, err := i.GetAddCurrency(stmt.CurDef.String())
	if err != nil {
		return err
	}

	account := Account{
		AccountId:         int64(len(i.Accounts) + 1),
		ExternalAccountId: stmt.BankAcctFrom.AcctID.String(),
		SecurityId:        security.SecurityId,
		ParentAccountId:   -1,
		Type:              Bank,
	}

	for _, tran := range stmt.BankTranList.Transactions {
		if err := i.AddTransaction(&tran, &account); err != nil {
			return err
		}
	}

	i.Accounts = append(i.Accounts, account)

	return nil
}

func (i *OFXImport) importOFXCC(stmt *ofxgo.CCStatementResponse) error {
	security, err := i.GetAddCurrency(stmt.CurDef.String())
	if err != nil {
		return err
	}

	account := Account{
		AccountId:         int64(len(i.Accounts) + 1),
		ExternalAccountId: stmt.CCAcctFrom.AcctID.String(),
		SecurityId:        security.SecurityId,
		ParentAccountId:   -1,
		Type:              Bank,
	}
	i.Accounts = append(i.Accounts, account)

	for _, tran := range stmt.BankTranList.Transactions {
		if err := i.AddTransaction(&tran, &account); err != nil {
			return err
		}
	}

	return nil
}

func (i *OFXImport) importOFXInv(stmt *ofxgo.InvStatementResponse) error {
	// TODO
	return errors.New("unimplemented")
}

func ImportOFX(r io.Reader) (*OFXImport, error) {
	var i OFXImport

	response, err := ofxgo.ParseResponse(r)
	if err != nil {
		return nil, fmt.Errorf("Unexpected error parsing OFX response: %s\n", err)
	}

	if response.Signon.Status.Code != 0 {
		meaning, _ := response.Signon.Status.CodeMeaning()
		return nil, fmt.Errorf("Nonzero signon status (%d: %s) with message: %s\n", response.Signon.Status.Code, meaning, response.Signon.Status.Message)
	}

	for _, bank := range response.Bank {
		if stmt, ok := bank.(*ofxgo.StatementResponse); ok {
			err = i.importOFXBank(stmt)
			if err != nil {
				return nil, err
			}
			return &i, nil
		}
	}
	for _, cc := range response.CreditCard {
		if stmt, ok := cc.(*ofxgo.CCStatementResponse); ok {
			err = i.importOFXCC(stmt)
			if err != nil {
				return nil, err
			}
			return &i, nil
		}
	}
	for _, inv := range response.InvStmt {
		if stmt, ok := inv.(*ofxgo.InvStatementResponse); ok {
			err = i.importOFXInv(stmt)
			if err != nil {
				return nil, err
			}
			return &i, nil
		}
	}

	return nil, errors.New("No OFX statement found")
}
