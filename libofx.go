package main

//#cgo LDFLAGS: -lofx
//
//#include <stdlib.h>
//
// //The next line disables the definition of static variables to allow for it to
// //be included here (see libofx commit bd24df15531e52a2858f70487443af8b9fa407f4)
//#define OFX_AQUAMANIAC_UGLY_HACK1
//#include <libofx/libofx.h>
//
// typedef int (*ofx_statement_cb_fn) (const struct OfxStatementData, void *);
// extern int ofx_statement_callback(const struct OfxStatementData, void *);
// typedef int (*ofx_account_cb_fn) (const struct OfxAccountData, void *);
// extern int ofx_account_callback(const struct OfxAccountData, void *);
// typedef int (*ofx_transaction_cb_fn) (const struct OfxTransactionData, void *);
// extern int ofx_transaction_callback(const struct OfxTransactionData, void *);
import "C"

import (
	"errors"
	"math/big"
	"time"
	"unsafe"
)

type ImportObject struct {
	TransactionList OFXImport
	Error           error
}

type OFXImport struct {
	Account           *Account
	Transactions      *[]Transaction
	TotalTransactions int64
	BeginningBalance  string
	EndingBalance     string
}

func init() {
	// Turn off all libofx info/debug messages
	C.ofx_PARSER_msg = 0
	C.ofx_DEBUG_msg = 0
	C.ofx_DEBUG1_msg = 0
	C.ofx_DEBUG2_msg = 0
	C.ofx_DEBUG3_msg = 0
	C.ofx_DEBUG4_msg = 0
	C.ofx_DEBUG5_msg = 0
	C.ofx_STATUS_msg = 0
	C.ofx_INFO_msg = 0
	C.ofx_WARNING_msg = 0
	C.ofx_ERROR_msg = 0
}

//export OFXStatementCallback
func OFXStatementCallback(statement_data C.struct_OfxStatementData, data unsafe.Pointer) C.int {
	//	import := (*ImportObject)(data)
	return 0
}

//export OFXAccountCallback
func OFXAccountCallback(account_data C.struct_OfxAccountData, data unsafe.Pointer) C.int {
	iobj := (*ImportObject)(data)
	itl := iobj.TransactionList
	if account_data.account_id_valid != 0 {
		account_name := C.GoString(&account_data.account_name[0])
		account_id := C.GoString(&account_data.account_id[0])
		itl.Account.Name = account_name
		itl.Account.ExternalAccountId = account_id
	} else {
		if iobj.Error == nil {
			iobj.Error = errors.New("OFX account ID invalid")
		}
		return 1
	}
	if account_data.account_type_valid != 0 {
		switch account_data.account_type {
		case C.OFX_CHECKING, C.OFX_SAVINGS, C.OFX_MONEYMRKT, C.OFX_CMA:
			itl.Account.Type = Bank
		case C.OFX_CREDITLINE, C.OFX_CREDITCARD:
			itl.Account.Type = Liability
		case C.OFX_INVESTMENT:
			itl.Account.Type = Investment
		}
	} else {
		if iobj.Error == nil {
			iobj.Error = errors.New("OFX account type invalid")
		}
		return 1
	}
	if account_data.currency_valid != 0 {
		currency_name := C.GoString(&account_data.currency[0])
		currency, err := GetSecurityByName(currency_name)
		if err != nil {
			if iobj.Error == nil {
				iobj.Error = err
			}
			return 1
		}
		itl.Account.SecurityId = currency.SecurityId
	} else {
		if iobj.Error == nil {
			iobj.Error = errors.New("OFX account currency invalid")
		}
		return 1
	}
	return 0
}

//export OFXTransactionCallback
func OFXTransactionCallback(transaction_data C.struct_OfxTransactionData, data unsafe.Pointer) C.int {
	iobj := (*ImportObject)(data)
	itl := iobj.TransactionList
	transaction := new(Transaction)

	if transaction_data.name_valid != 0 {
		transaction.Description = C.GoString(&transaction_data.name[0])
	}
	//	if transaction_data.reference_number_valid != 0 {
	//		fmt.Println("reference_number: ", C.GoString(&transaction_data.reference_number[0]))
	//	}
	if transaction_data.date_posted_valid != 0 {
		transaction.Date = time.Unix(int64(transaction_data.date_posted), 0)
	} else if transaction_data.date_initiated_valid != 0 {
		transaction.Date = time.Unix(int64(transaction_data.date_initiated), 0)
	}
	if transaction_data.fi_id_valid != 0 {
		transaction.RemoteId = C.GoString(&transaction_data.fi_id[0])
	}

	if transaction_data.amount_valid != 0 {
		split := new(Split)
		r := new(big.Rat)
		r.SetFloat64(float64(transaction_data.amount))
		security := GetSecurity(itl.Account.SecurityId)
		split.Amount = r.FloatString(security.Precision)
		if transaction_data.memo_valid != 0 {
			split.Memo = C.GoString(&transaction_data.memo[0])
		}
		if transaction_data.check_number_valid != 0 {
			split.Number = C.GoString(&transaction_data.check_number[0])
		}
		split.SecurityId = -1
		split.AccountId = itl.Account.AccountId
		transaction.Splits = append(transaction.Splits, split)
	} else {
		if iobj.Error == nil {
			iobj.Error = errors.New("OFX transaction amount invalid")
		}
		return 1
	}

	var security *Security
	split := new(Split)
	units := new(big.Rat)

	if transaction_data.units_valid != 0 {
		units.SetFloat64(float64(transaction_data.units))
		if transaction_data.security_data_valid != 0 {
			security_data := transaction_data.security_data_ptr
			if security_data.ticker_valid != 0 {
				s, err := GetSecurityByName(C.GoString(&security_data.ticker[0]))
				if err != nil {
					if iobj.Error == nil {
						iobj.Error = errors.New("Failed to find OFX transaction security: " + C.GoString(&security_data.ticker[0]))
					}
					return 1
				}
				security = s
			} else {
				if iobj.Error == nil {
					iobj.Error = errors.New("OFX security ticker invalid")
				}
				return 1
			}
			if security.Type == Stock && security_data.unique_id_valid != 0 && security_data.unique_id_type_valid != 0 && C.GoString(&security_data.unique_id_type[0]) == "CUSIP" {
				// Validate the security CUSIP, if possible
				if security.AlternateId != C.GoString(&security_data.unique_id[0]) {
					if iobj.Error == nil {
						iobj.Error = errors.New("OFX transaction security CUSIP failed to validate")
					}
					return 1
				}
			}
		} else {
			security = GetSecurity(itl.Account.SecurityId)
		}
	} else {
		// Calculate units from other available fields if its not present
		//		units = - (amount + various fees) / unitprice
		units.SetFloat64(float64(transaction_data.amount))
		fees := new(big.Rat)
		if transaction_data.fees_valid != 0 {
			fees.SetFloat64(float64(-transaction_data.fees))
		}
		if transaction_data.commission_valid != 0 {
			commission := new(big.Rat)
			commission.SetFloat64(float64(-transaction_data.commission))
			fees.Add(fees, commission)
		}
		units.Add(units, fees)
		units.Neg(units)
		if transaction_data.unitprice_valid != 0 && transaction_data.unitprice != 0 {
			unitprice := new(big.Rat)
			unitprice.SetFloat64(float64(transaction_data.unitprice))
			units.Quo(units, unitprice)
		}

		// If 'units' wasn't present, assume we're using the account's security
		security = GetSecurity(itl.Account.SecurityId)
	}

	split.Amount = units.FloatString(security.Precision)
	split.SecurityId = security.SecurityId
	split.AccountId = -1
	transaction.Splits = append(transaction.Splits, split)

	if transaction_data.fees_valid != 0 {
		split := new(Split)
		r := new(big.Rat)
		r.SetFloat64(float64(-transaction_data.fees))
		security := GetSecurity(itl.Account.SecurityId)
		split.Amount = r.FloatString(security.Precision)
		split.Memo = "fees"
		split.SecurityId = itl.Account.SecurityId
		split.AccountId = -1
		transaction.Splits = append(transaction.Splits, split)
	}

	if transaction_data.commission_valid != 0 {
		split := new(Split)
		r := new(big.Rat)
		r.SetFloat64(float64(-transaction_data.commission))
		security := GetSecurity(itl.Account.SecurityId)
		split.Amount = r.FloatString(security.Precision)
		split.Memo = "commission"
		split.SecurityId = itl.Account.SecurityId
		split.AccountId = -1
		transaction.Splits = append(transaction.Splits, split)
	}

	//	if transaction_data.payee_id_valid != 0 {
	//		fmt.Println("payee_id: ", C.GoString(&transaction_data.payee_id[0]))
	//	}

	transaction_list := append(*itl.Transactions, *transaction)
	iobj.TransactionList.Transactions = &transaction_list

	return 0
}

func ImportOFX(filename string, account *Account) (*OFXImport, error) {
	var a Account
	var t []Transaction
	var iobj ImportObject
	iobj.TransactionList.Account = &a
	iobj.TransactionList.Transactions = &t

	a.AccountId = account.AccountId

	context := C.libofx_get_new_context()
	defer C.libofx_free_context(context)

	C.ofx_set_statement_cb(context, C.ofx_statement_cb_fn(C.ofx_statement_callback), unsafe.Pointer(&iobj))
	C.ofx_set_account_cb(context, C.ofx_account_cb_fn(C.ofx_account_callback), unsafe.Pointer(&iobj))
	C.ofx_set_transaction_cb(context, C.ofx_transaction_cb_fn(C.ofx_transaction_callback), unsafe.Pointer(&iobj))

	filename_cstring := C.CString(filename)
	defer C.free(unsafe.Pointer(filename_cstring))
	C.libofx_proc_file(context, filename_cstring, C.OFX) // unconditionally returns 0.

	iobj.TransactionList.TotalTransactions = int64(len(*iobj.TransactionList.Transactions))

	if iobj.TransactionList.TotalTransactions == 0 {
		return nil, errors.New("No OFX transactions found")
	}

	if iobj.Error != nil {
		return nil, iobj.Error
	} else {
		return &iobj.TransactionList, nil
	}
}
