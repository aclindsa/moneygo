package handlers

import (
	"errors"
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/ofxgo"
	"io"
	"math/big"
)

type OFXImport struct {
	Securities   []models.Security
	Accounts     []models.Account
	Transactions []models.Transaction
	//	Balances     map[int64]string // map AccountIDs to ending balances
}

func (i *OFXImport) GetSecurity(ofxsecurityid int64) (*models.Security, error) {
	if ofxsecurityid < 0 || ofxsecurityid > int64(len(i.Securities)) {
		return nil, errors.New("OFXImport.GetSecurity: SecurityID out of range")
	}
	return &i.Securities[ofxsecurityid], nil
}

func (i *OFXImport) GetSecurityAlternateId(alternateid string, securityType models.SecurityType) (*models.Security, error) {
	for _, security := range i.Securities {
		if alternateid == security.AlternateId && securityType == security.Type {
			return &security, nil
		}
	}

	return nil, errors.New("OFXImport.FindSecurity: Unable to find security")
}

func (i *OFXImport) GetAddCurrency(isoname string) (*models.Security, error) {
	for _, security := range i.Securities {
		if isoname == security.Name && models.Currency == security.Type {
			return &security, nil
		}
	}

	template := FindSecurityTemplate(isoname, models.Currency)
	if template == nil {
		return nil, fmt.Errorf("Failed to find Security for \"%s\"", isoname)
	}
	var security models.Security = *template
	security.SecurityId = int64(len(i.Securities) + 1)
	i.Securities = append(i.Securities, security)

	return &security, nil
}

func (i *OFXImport) AddTransaction(tran *ofxgo.Transaction, account *models.Account) error {
	var t models.Transaction

	t.Date = tran.DtPosted.UTC()

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

	var s1, s2 models.Split
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

	s1.RemoteId = "ofx:" + tran.FiTID.String()
	// TODO CorrectFiTID/CorrectAction?

	s1.ImportSplitType = models.ImportAccount
	s2.ImportSplitType = models.ExternalAccount

	s1.Amount.Rat = *amt
	s2.Amount.Rat = *amt.Neg(amt)
	security := i.Securities[account.SecurityId-1]
	if s1.Amount.Precision() > security.Precision {
		return errors.New("Imported transaction amount is too precise for security")
	}

	s1.Status = models.Imported
	s2.Status = models.Imported

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

	account := models.Account{
		AccountId:         int64(len(i.Accounts) + 1),
		ExternalAccountId: stmt.BankAcctFrom.AcctID.String(),
		SecurityId:        security.SecurityId,
		ParentAccountId:   -1,
		Type:              models.Bank,
	}

	if stmt.BankTranList != nil {
		for _, tran := range stmt.BankTranList.Transactions {
			if err := i.AddTransaction(&tran, &account); err != nil {
				return err
			}
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

	account := models.Account{
		AccountId:         int64(len(i.Accounts) + 1),
		ExternalAccountId: stmt.CCAcctFrom.AcctID.String(),
		SecurityId:        security.SecurityId,
		ParentAccountId:   -1,
		Type:              models.Liability,
	}
	i.Accounts = append(i.Accounts, account)

	if stmt.BankTranList != nil {
		for _, tran := range stmt.BankTranList.Transactions {
			if err := i.AddTransaction(&tran, &account); err != nil {
				return err
			}
		}
	}

	// TODO balance(s)

	return nil
}

func (i *OFXImport) importSecurities(seclist *ofxgo.SecurityList) error {
	for _, security := range seclist.Securities {
		var si ofxgo.SecInfo
		if sec, ok := (security).(ofxgo.DebtInfo); ok {
			si = sec.SecInfo
		} else if sec, ok := (security).(ofxgo.MFInfo); ok {
			si = sec.SecInfo
		} else if sec, ok := (security).(ofxgo.OptInfo); ok {
			si = sec.SecInfo
		} else if sec, ok := (security).(ofxgo.OtherInfo); ok {
			si = sec.SecInfo
		} else if sec, ok := (security).(ofxgo.StockInfo); ok {
			si = sec.SecInfo
		} else {
			return errors.New("Can't import unrecognized type satisfying ofxgo.Security interface")
		}
		s := models.Security{
			SecurityId:  int64(len(i.Securities) + 1),
			Name:        string(si.SecName),
			Description: string(si.Memo),
			Symbol:      string(si.Ticker),
			Precision:   5, // TODO How to actually determine this?
			Type:        models.Stock,
			AlternateId: string(si.SecID.UniqueID),
		}
		if len(s.Description) == 0 {
			s.Description = s.Name
		}
		if len(s.Symbol) == 0 {
			s.Symbol = s.Name
		}

		i.Securities = append(i.Securities, s)
	}
	return nil
}

func (i *OFXImport) GetInvTran(invtran *ofxgo.InvTran) models.Transaction {
	var t models.Transaction
	t.Description = string(invtran.Memo)
	t.Date = invtran.DtTrade.UTC()
	return t
}

func (i *OFXImport) GetInvBuyTran(buy *ofxgo.InvBuy, curdef *models.Security, account *models.Account) (*models.Transaction, error) {
	t := i.GetInvTran(&buy.InvTran)

	security, err := i.GetSecurityAlternateId(string(buy.SecID.UniqueID), models.Stock)
	if err != nil {
		return nil, err
	}

	memo := string(buy.InvTran.Memo)
	if len(memo) > 0 {
		memo += " "
	}

	var commission, taxes, fees, load, total, tradingTotal big.Rat
	commission.Abs(&buy.Commission.Rat)
	taxes.Abs(&buy.Taxes.Rat)
	fees.Abs(&buy.Fees.Rat)
	load.Abs(&buy.Load.Rat)
	total.Abs(&buy.Total.Rat)

	total.Neg(&total)

	tradingTotal.Neg(&total)
	tradingTotal.Sub(&tradingTotal, &commission)
	tradingTotal.Sub(&tradingTotal, &taxes)
	tradingTotal.Sub(&tradingTotal, &fees)
	tradingTotal.Sub(&tradingTotal, &load)

	// Convert amounts to account's currency if Currency is set
	if ok, _ := buy.Currency.Valid(); ok {
		commission.Mul(&commission, &buy.Currency.CurRate.Rat)
		taxes.Mul(&taxes, &buy.Currency.CurRate.Rat)
		fees.Mul(&fees, &buy.Currency.CurRate.Rat)
		load.Mul(&load, &buy.Currency.CurRate.Rat)
		total.Mul(&total, &buy.Currency.CurRate.Rat)
		tradingTotal.Mul(&tradingTotal, &buy.Currency.CurRate.Rat)
	}

	if num := commission.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Commission,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + buy.InvTran.FiTID.String(),
			Memo:            memo + "(commission)",
			Amount:          models.Amount{commission},
		})
	}
	if num := taxes.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Taxes,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + buy.InvTran.FiTID.String(),
			Memo:            memo + "(taxes)",
			Amount:          models.Amount{taxes},
		})
	}
	if num := fees.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Fees,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + buy.InvTran.FiTID.String(),
			Memo:            memo + "(fees)",
			Amount:          models.Amount{fees},
		})
	}
	if num := load.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Load,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + buy.InvTran.FiTID.String(),
			Memo:            memo + "(load)",
			Amount:          models.Amount{load},
		})
	}
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.ImportAccount,
		AccountId:       account.AccountId,
		SecurityId:      -1,
		RemoteId:        "ofx:" + buy.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.TradingAccount,
		AccountId:       -1,
		SecurityId:      curdef.SecurityId,
		RemoteId:        "ofx:" + buy.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{tradingTotal},
	})

	var units big.Rat
	units.Abs(&buy.Units.Rat)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.SubAccount,
		AccountId:       -1,
		SecurityId:      security.SecurityId,
		RemoteId:        "ofx:" + buy.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{units},
	})
	units.Neg(&units)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.TradingAccount,
		AccountId:       -1,
		SecurityId:      security.SecurityId,
		RemoteId:        "ofx:" + buy.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{units},
	})

	return &t, nil
}

func (i *OFXImport) GetIncomeTran(income *ofxgo.Income, curdef *models.Security, account *models.Account) (*models.Transaction, error) {
	t := i.GetInvTran(&income.InvTran)

	security, err := i.GetSecurityAlternateId(string(income.SecID.UniqueID), models.Stock)
	if err != nil {
		return nil, err
	}

	memo := string(income.InvTran.Memo)
	if len(memo) > 0 {
		memo += " "
	} else {
		memo = income.IncomeType.String() + " on " + security.Symbol
	}

	var total big.Rat
	total.Set(&income.Total.Rat)
	if ok, _ := income.Currency.Valid(); ok {
		total.Mul(&total, &income.Currency.CurRate.Rat)
	}

	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.ImportAccount,
		AccountId:       account.AccountId,
		SecurityId:      -1,
		RemoteId:        "ofx:" + income.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})
	total.Neg(&total)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.IncomeAccount,
		AccountId:       -1,
		SecurityId:      curdef.SecurityId,
		RemoteId:        "ofx:" + income.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})

	return &t, nil
}

func (i *OFXImport) GetInvExpenseTran(expense *ofxgo.InvExpense, curdef *models.Security, account *models.Account) (*models.Transaction, error) {
	t := i.GetInvTran(&expense.InvTran)

	security, err := i.GetSecurityAlternateId(string(expense.SecID.UniqueID), models.Stock)
	if err != nil {
		return nil, err
	}

	memo := string(expense.InvTran.Memo)
	if len(memo) == 0 {
		memo = "INVEXPENSE"
	}
	memo += " (" + security.Symbol + ")"

	var total big.Rat
	total.Set(&expense.Total.Rat)
	if ok, _ := expense.Currency.Valid(); ok {
		total.Mul(&total, &expense.Currency.CurRate.Rat)
	}

	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.ImportAccount,
		AccountId:       account.AccountId,
		SecurityId:      -1,
		RemoteId:        "ofx:" + expense.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})
	total.Neg(&total)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.ExpenseAccount,
		AccountId:       -1,
		SecurityId:      curdef.SecurityId,
		RemoteId:        "ofx:" + expense.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})

	return &t, nil
}

func (i *OFXImport) GetMarginInterestTran(marginint *ofxgo.MarginInterest, curdef *models.Security, account *models.Account) (*models.Transaction, error) {
	t := i.GetInvTran(&marginint.InvTran)

	memo := string(marginint.InvTran.Memo)
	if len(memo) == 0 {
		memo = "MARGININTEREST"
	}

	var total big.Rat
	total.Set(&marginint.Total.Rat)
	if ok, _ := marginint.Currency.Valid(); ok {
		total.Mul(&total, &marginint.Currency.CurRate.Rat)
	}

	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.ImportAccount,
		AccountId:       account.AccountId,
		SecurityId:      -1,
		RemoteId:        "ofx:" + marginint.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})
	total.Neg(&total)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.IncomeAccount,
		AccountId:       -1,
		SecurityId:      curdef.SecurityId,
		RemoteId:        "ofx:" + marginint.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})

	return &t, nil
}

func (i *OFXImport) GetReinvestTran(reinvest *ofxgo.Reinvest, curdef *models.Security, account *models.Account) (*models.Transaction, error) {
	t := i.GetInvTran(&reinvest.InvTran)

	security, err := i.GetSecurityAlternateId(string(reinvest.SecID.UniqueID), models.Stock)
	if err != nil {
		return nil, err
	}

	memo := string(reinvest.InvTran.Memo)
	if len(memo) > 0 {
		memo += " "
	}

	var commission, taxes, fees, load, total, tradingTotal big.Rat
	commission.Abs(&reinvest.Commission.Rat)
	taxes.Abs(&reinvest.Taxes.Rat)
	fees.Abs(&reinvest.Fees.Rat)
	load.Abs(&reinvest.Load.Rat)
	total.Abs(&reinvest.Total.Rat)

	total.Neg(&total)

	tradingTotal.Neg(&total)
	tradingTotal.Sub(&tradingTotal, &commission)
	tradingTotal.Sub(&tradingTotal, &taxes)
	tradingTotal.Sub(&tradingTotal, &fees)
	tradingTotal.Sub(&tradingTotal, &load)

	// Convert amounts to account's currency if Currency is set
	if ok, _ := reinvest.Currency.Valid(); ok {
		commission.Mul(&commission, &reinvest.Currency.CurRate.Rat)
		taxes.Mul(&taxes, &reinvest.Currency.CurRate.Rat)
		fees.Mul(&fees, &reinvest.Currency.CurRate.Rat)
		load.Mul(&load, &reinvest.Currency.CurRate.Rat)
		total.Mul(&total, &reinvest.Currency.CurRate.Rat)
		tradingTotal.Mul(&tradingTotal, &reinvest.Currency.CurRate.Rat)
	}

	if num := commission.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Commission,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + reinvest.InvTran.FiTID.String(),
			Memo:            memo + "(commission)",
			Amount:          models.Amount{commission},
		})
	}
	if num := taxes.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Taxes,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + reinvest.InvTran.FiTID.String(),
			Memo:            memo + "(taxes)",
			Amount:          models.Amount{taxes},
		})
	}
	if num := fees.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Fees,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + reinvest.InvTran.FiTID.String(),
			Memo:            memo + "(fees)",
			Amount:          models.Amount{fees},
		})
	}
	if num := load.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Load,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + reinvest.InvTran.FiTID.String(),
			Memo:            memo + "(load)",
			Amount:          models.Amount{load},
		})
	}
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.ImportAccount,
		AccountId:       account.AccountId,
		SecurityId:      -1,
		RemoteId:        "ofx:" + reinvest.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})

	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.IncomeAccount,
		AccountId:       -1,
		SecurityId:      curdef.SecurityId,
		RemoteId:        "ofx:" + reinvest.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})
	total.Neg(&total)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.ImportAccount,
		AccountId:       account.AccountId,
		SecurityId:      -1,
		RemoteId:        "ofx:" + reinvest.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.TradingAccount,
		AccountId:       -1,
		SecurityId:      curdef.SecurityId,
		RemoteId:        "ofx:" + reinvest.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{tradingTotal},
	})

	var units big.Rat
	units.Abs(&reinvest.Units.Rat)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.SubAccount,
		AccountId:       -1,
		SecurityId:      security.SecurityId,
		RemoteId:        "ofx:" + reinvest.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{units},
	})
	units.Neg(&units)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.TradingAccount,
		AccountId:       -1,
		SecurityId:      security.SecurityId,
		RemoteId:        "ofx:" + reinvest.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{units},
	})

	return &t, nil
}

func (i *OFXImport) GetRetOfCapTran(retofcap *ofxgo.RetOfCap, curdef *models.Security, account *models.Account) (*models.Transaction, error) {
	t := i.GetInvTran(&retofcap.InvTran)

	security, err := i.GetSecurityAlternateId(string(retofcap.SecID.UniqueID), models.Stock)
	if err != nil {
		return nil, err
	}

	memo := string(retofcap.InvTran.Memo)
	if len(memo) == 0 {
		memo = "RETOFCAP"
	}
	memo += " (" + security.Symbol + ")"

	var total big.Rat
	total.Set(&retofcap.Total.Rat)
	if ok, _ := retofcap.Currency.Valid(); ok {
		total.Mul(&total, &retofcap.Currency.CurRate.Rat)
	}

	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.ImportAccount,
		AccountId:       account.AccountId,
		SecurityId:      -1,
		RemoteId:        "ofx:" + retofcap.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})
	total.Neg(&total)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.IncomeAccount,
		AccountId:       -1,
		SecurityId:      curdef.SecurityId,
		RemoteId:        "ofx:" + retofcap.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})

	return &t, nil
}

func (i *OFXImport) GetInvSellTran(sell *ofxgo.InvSell, curdef *models.Security, account *models.Account) (*models.Transaction, error) {
	t := i.GetInvTran(&sell.InvTran)

	security, err := i.GetSecurityAlternateId(string(sell.SecID.UniqueID), models.Stock)
	if err != nil {
		return nil, err
	}

	memo := string(sell.InvTran.Memo)
	if len(memo) > 0 {
		memo += " "
	}

	var commission, taxes, fees, load, total, tradingTotal big.Rat
	commission.Abs(&sell.Commission.Rat)
	taxes.Abs(&sell.Taxes.Rat)
	fees.Abs(&sell.Fees.Rat)
	load.Abs(&sell.Load.Rat)
	total.Abs(&sell.Total.Rat)

	commission.Neg(&commission)
	taxes.Neg(&taxes)
	fees.Neg(&fees)
	load.Neg(&load)

	tradingTotal.Neg(&total)
	tradingTotal.Add(&tradingTotal, &commission)
	tradingTotal.Add(&tradingTotal, &taxes)
	tradingTotal.Add(&tradingTotal, &fees)
	tradingTotal.Add(&tradingTotal, &load)

	// Convert amounts to account's currency if Currency is set
	if ok, _ := sell.Currency.Valid(); ok {
		commission.Mul(&commission, &sell.Currency.CurRate.Rat)
		taxes.Mul(&taxes, &sell.Currency.CurRate.Rat)
		fees.Mul(&fees, &sell.Currency.CurRate.Rat)
		load.Mul(&load, &sell.Currency.CurRate.Rat)
		total.Mul(&total, &sell.Currency.CurRate.Rat)
		tradingTotal.Mul(&tradingTotal, &sell.Currency.CurRate.Rat)
	}

	if num := commission.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Commission,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + sell.InvTran.FiTID.String(),
			Memo:            memo + "(commission)",
			Amount:          models.Amount{commission},
		})
	}
	if num := taxes.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Taxes,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + sell.InvTran.FiTID.String(),
			Memo:            memo + "(taxes)",
			Amount:          models.Amount{taxes},
		})
	}
	if num := fees.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Fees,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + sell.InvTran.FiTID.String(),
			Memo:            memo + "(fees)",
			Amount:          models.Amount{fees},
		})
	}
	if num := load.Num(); !num.IsInt64() || num.Int64() != 0 {
		t.Splits = append(t.Splits, &models.Split{
			// TODO ReversalFiTID?
			Status:          models.Imported,
			ImportSplitType: models.Load,
			AccountId:       -1,
			SecurityId:      curdef.SecurityId,
			RemoteId:        "ofx:" + sell.InvTran.FiTID.String(),
			Memo:            memo + "(load)",
			Amount:          models.Amount{load},
		})
	}
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.ImportAccount,
		AccountId:       account.AccountId,
		SecurityId:      -1,
		RemoteId:        "ofx:" + sell.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{total},
	})
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.TradingAccount,
		AccountId:       -1,
		SecurityId:      curdef.SecurityId,
		RemoteId:        "ofx:" + sell.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{tradingTotal},
	})

	var units big.Rat
	units.Abs(&sell.Units.Rat)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.TradingAccount,
		AccountId:       -1,
		SecurityId:      security.SecurityId,
		RemoteId:        "ofx:" + sell.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{units},
	})
	units.Neg(&units)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.SubAccount,
		AccountId:       -1,
		SecurityId:      security.SecurityId,
		RemoteId:        "ofx:" + sell.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{units},
	})

	return &t, nil
}

func (i *OFXImport) GetTransferTran(transfer *ofxgo.Transfer, account *models.Account) (*models.Transaction, error) {
	t := i.GetInvTran(&transfer.InvTran)

	security, err := i.GetSecurityAlternateId(string(transfer.SecID.UniqueID), models.Stock)
	if err != nil {
		return nil, err
	}

	memo := string(transfer.InvTran.Memo)

	var units big.Rat
	if transfer.TferAction == ofxgo.TferActionIn {
		units.Set(&transfer.Units.Rat)
	} else {
		units.Neg(&transfer.Units.Rat)
	}

	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.SubAccount,
		AccountId:       -1,
		SecurityId:      security.SecurityId,
		RemoteId:        "ofx:" + transfer.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{units},
	})
	units.Neg(&units)
	t.Splits = append(t.Splits, &models.Split{
		// TODO ReversalFiTID?
		Status:          models.Imported,
		ImportSplitType: models.ExternalAccount,
		AccountId:       -1,
		SecurityId:      security.SecurityId,
		RemoteId:        "ofx:" + transfer.InvTran.FiTID.String(),
		Memo:            memo,
		Amount:          models.Amount{units},
	})

	return &t, nil
}

func (i *OFXImport) AddInvTransaction(invtran *ofxgo.InvTransaction, account *models.Account, curdef *models.Security) error {
	if curdef.SecurityId < 1 || curdef.SecurityId > int64(len(i.Securities)) {
		return errors.New("Internal error: security index not found in OFX import\n")
	}

	var t *models.Transaction
	var err error
	if tran, ok := (*invtran).(ofxgo.BuyDebt); ok {
		t, err = i.GetInvBuyTran(&tran.InvBuy, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.BuyMF); ok {
		t, err = i.GetInvBuyTran(&tran.InvBuy, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.BuyOpt); ok {
		t, err = i.GetInvBuyTran(&tran.InvBuy, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.BuyOther); ok {
		t, err = i.GetInvBuyTran(&tran.InvBuy, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.BuyStock); ok {
		t, err = i.GetInvBuyTran(&tran.InvBuy, curdef, account)
		//	} else if tran, ok := (*invtran).(ofxgo.ClosureOpt); ok {
		// TODO implementme
	} else if tran, ok := (*invtran).(ofxgo.Income); ok {
		t, err = i.GetIncomeTran(&tran, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.InvExpense); ok {
		t, err = i.GetInvExpenseTran(&tran, curdef, account)
		//	} else if tran, ok := (*invtran).(ofxgo.JrnlFund); ok {
		// TODO implementme
		//	} else if tran, ok := (*invtran).(ofxgo.JrnlSec); ok {
		// TODO implementme
	} else if tran, ok := (*invtran).(ofxgo.MarginInterest); ok {
		t, err = i.GetMarginInterestTran(&tran, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.Reinvest); ok {
		t, err = i.GetReinvestTran(&tran, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.RetOfCap); ok {
		t, err = i.GetRetOfCapTran(&tran, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.SellDebt); ok {
		t, err = i.GetInvSellTran(&tran.InvSell, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.SellMF); ok {
		t, err = i.GetInvSellTran(&tran.InvSell, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.SellOpt); ok {
		t, err = i.GetInvSellTran(&tran.InvSell, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.SellOther); ok {
		t, err = i.GetInvSellTran(&tran.InvSell, curdef, account)
	} else if tran, ok := (*invtran).(ofxgo.SellStock); ok {
		t, err = i.GetInvSellTran(&tran.InvSell, curdef, account)
		//	} else if tran, ok := (*invtran).(ofxgo.Split); ok {
		// TODO implementme
	} else if tran, ok := (*invtran).(ofxgo.Transfer); ok {
		t, err = i.GetTransferTran(&tran, account)
	} else {
		return errors.New("Unrecognized type satisfying ofxgo.InvTransaction interface: " + (*invtran).TransactionType())

	}

	if err != nil {
		return err
	}

	i.Transactions = append(i.Transactions, *t)

	return nil
}

func (i *OFXImport) importOFXInv(stmt *ofxgo.InvStatementResponse) error {
	security, err := i.GetAddCurrency(stmt.CurDef.String())
	if err != nil {
		return err
	}

	account := models.Account{
		AccountId:         int64(len(i.Accounts) + 1),
		ExternalAccountId: stmt.InvAcctFrom.AcctID.String(),
		SecurityId:        security.SecurityId,
		ParentAccountId:   -1,
		Type:              models.Investment,
	}
	i.Accounts = append(i.Accounts, account)

	if stmt.InvTranList != nil {
		for _, invtran := range stmt.InvTranList.InvTransactions {
			if err := i.AddInvTransaction(&invtran, &account, security); err != nil {
				return err
			}
		}
		for _, bt := range stmt.InvTranList.BankTransactions {
			// TODO Should we do something different for the value of
			// bt.SubAcctFund?
			for _, tran := range bt.Transactions {
				if err := i.AddTransaction(&tran, &account); err != nil {
					return err
				}
			}
		}
	}

	// TODO InvPosList
	// TODO InvBal
	// TODO Inv401K and INV401kBal???

	return nil
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
	for _, seclist := range response.SecList {
		if securitylist, ok := seclist.(*ofxgo.SecurityList); ok {
			err = i.importSecurities(securitylist)
			if err != nil {
				return nil, err
			}
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
