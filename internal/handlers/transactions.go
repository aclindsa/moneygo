package handlers

import (
	"errors"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
)

// Return a map of security ID's to big.Rat's containing the amount that
// security is imbalanced by
func GetTransactionImbalances(tx store.Tx, t *models.Transaction) (map[int64]big.Rat, error) {
	sums := make(map[int64]big.Rat)

	if !t.Valid() {
		return nil, errors.New("Transaction invalid")
	}

	for i := range t.Splits {
		securityid := t.Splits[i].SecurityId
		if t.Splits[i].AccountId != -1 {
			var err error
			var account *models.Account
			account, err = tx.GetAccount(t.Splits[i].AccountId, t.UserId)
			if err != nil {
				return nil, err
			}
			securityid = account.SecurityId
		}
		sum := sums[securityid]
		(&sum).Add(&sum, &t.Splits[i].Amount.Rat)
		sums[securityid] = sum
	}
	return sums, nil
}

// Returns true if all securities contained in this transaction are balanced,
// false otherwise
func TransactionBalanced(tx store.Tx, t *models.Transaction) (bool, error) {
	var zero big.Rat

	sums, err := GetTransactionImbalances(tx, t)
	if err != nil {
		return false, err
	}

	for _, security_sum := range sums {
		if security_sum.Cmp(&zero) != 0 {
			return false, nil
		}
	}
	return true, nil
}

func TransactionHandler(r *http.Request, context *Context) ResponseWriterWriter {
	user, err := GetUserFromSession(context.Tx, r)
	if err != nil {
		return NewError(1 /*Not Signed In*/)
	}

	if r.Method == "POST" {
		var transaction models.Transaction
		if err := ReadJSON(r, &transaction); err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		transaction.TransactionId = -1
		transaction.UserId = user.UserId

		if len(transaction.Splits) == 0 {
			return NewError(3 /*Invalid Request*/)
		}

		for i := range transaction.Splits {
			transaction.Splits[i].SplitId = -1
			_, err := context.Tx.GetAccount(transaction.Splits[i].AccountId, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}
		}

		balanced, err := TransactionBalanced(context.Tx, &transaction)
		if err != nil {
			return NewError(999 /*Internal Error*/)
		}
		if !transaction.Valid() || !balanced {
			return NewError(3 /*Invalid Request*/)
		}

		err = context.Tx.InsertTransaction(&transaction, user)
		if err != nil {
			if _, ok := err.(store.AccountMissingError); ok {
				return NewError(3 /*Invalid Request*/)
			} else {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
		}

		return &transaction
	} else if r.Method == "GET" {
		if context.LastLevel() {
			//Return all Transactions
			var al models.TransactionList
			transactions, err := context.Tx.GetTransactions(user.UserId)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
			al.Transactions = transactions
			return &al
		} else {
			//Return Transaction with this Id
			transactionid, err := context.NextID()
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}
			transaction, err := context.Tx.GetTransaction(transactionid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}
			return transaction
		}
	} else {
		transactionid, err := context.NextID()
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		if r.Method == "PUT" {
			var transaction models.Transaction
			if err := ReadJSON(r, &transaction); err != nil || transaction.TransactionId != transactionid {
				return NewError(3 /*Invalid Request*/)
			}
			transaction.UserId = user.UserId

			balanced, err := TransactionBalanced(context.Tx, &transaction)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
			if !transaction.Valid() || !balanced {
				return NewError(3 /*Invalid Request*/)
			}

			if len(transaction.Splits) == 0 {
				return NewError(3 /*Invalid Request*/)
			}

			for i := range transaction.Splits {
				_, err := context.Tx.GetAccount(transaction.Splits[i].AccountId, user.UserId)
				if err != nil {
					return NewError(3 /*Invalid Request*/)
				}
			}

			err = context.Tx.UpdateTransaction(&transaction, user)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return &transaction
		} else if r.Method == "DELETE" {
			transaction, err := context.Tx.GetTransaction(transactionid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			err = context.Tx.DeleteTransaction(transaction, user)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}

// Return only those transactions which have at least one split pertaining to
// an account
func AccountTransactionsHandler(context *Context, r *http.Request, user *models.User, accountid int64) ResponseWriterWriter {
	var page uint64 = 0
	var limit uint64 = 50
	var sort string = "date-desc"

	query, _ := url.ParseQuery(r.URL.RawQuery)

	pagestring := query.Get("page")
	if pagestring != "" {
		p, err := strconv.ParseUint(pagestring, 10, 0)
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		page = p
	}

	limitstring := query.Get("limit")
	if limitstring != "" {
		l, err := strconv.ParseUint(limitstring, 10, 0)
		if err != nil || l > 100 {
			return NewError(3 /*Invalid Request*/)
		}
		limit = l
	}

	sortstring := query.Get("sort")
	if sortstring != "" {
		if sortstring != "date-asc" && sortstring != "date-desc" {
			return NewError(3 /*Invalid Request*/)
		}
		sort = sortstring
	}

	accountTransactions, err := context.Tx.GetAccountTransactions(user, accountid, sort, page, limit)
	if err != nil {
		log.Print(err)
		return NewError(999 /*Internal Error*/)
	}

	return accountTransactions
}
