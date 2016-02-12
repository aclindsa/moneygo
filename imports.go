package main

import (
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
)

/*
 * Assumes the User is a valid, signed-in user, but accountid has not yet been validated
 */
func AccountImportHandler(w http.ResponseWriter, r *http.Request, user *User, accountid int64) {
	// Return Account with this Id
	account, err := GetAccount(accountid, user.UserId)
	if err != nil {
		WriteError(w, 3 /*Invalid Request*/)
		return
	}

	multipartReader, err := r.MultipartReader()
	if err != nil {
		WriteError(w, 3 /*Invalid Request*/)
		return
	}

	// assume there is only one 'part'
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

	f, err := ioutil.TempFile(tmpDir, user.Username+"_"+account.Name)
	if err != nil {
		WriteError(w, 999 /*Internal Error*/)
		log.Print(err)
		return
	}
	tmpFilename := f.Name()
	defer os.Remove(tmpFilename)

	_, err = io.Copy(f, part)
	f.Close()
	if err != nil {
		WriteError(w, 999 /*Internal Error*/)
		log.Print(err)
		return
	}

	itl, err := ImportOFX(tmpFilename, account)

	if err != nil {
		//TODO is this necessarily an invalid request?
		WriteError(w, 3 /*Invalid Request*/)
		return
	}

	var transactions []Transaction
	for _, transaction := range *itl.Transactions {
		transaction.UserId = user.UserId
		transaction.Status = Imported

		if !transaction.Valid() {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		imbalances, err := transaction.GetImbalances()
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}

		// Fixup any imbalances in transactions
		var zero big.Rat
		var num_imbalances int
		for _, imbalance := range imbalances {
			if imbalance.Cmp(&zero) != 0 {
				num_imbalances += 1
			}
		}

		for imbalanced_security, imbalance := range imbalances {
			if imbalance.Cmp(&zero) != 0 {
				var imbalanced_account *Account
				// If we're dealing with exactly two securities, assume any imbalances
				// from imports are from trading currencies/securities
				if num_imbalances == 2 {
					imbalanced_account, err = GetTradingAccount(user.UserId, imbalanced_security)
				} else {
					imbalanced_account, err = GetImbalanceAccount(user.UserId, imbalanced_security)
				}
				if err != nil {
					WriteError(w, 999 /*Internal Error*/)
					log.Print(err)
					return
				}

				// Add new split to fixup imbalance
				split := new(Split)
				r := new(big.Rat)
				r.Neg(&imbalance)
				security := GetSecurity(imbalanced_security)
				split.Amount = r.FloatString(security.Precision)
				split.SecurityId = -1
				split.AccountId = imbalanced_account.AccountId
				transaction.Splits = append(transaction.Splits, split)
			}
		}

		// Move any splits with SecurityId but not AccountId to Imbalances
		// accounts
		for _, split := range transaction.Splits {
			if split.SecurityId != -1 || split.AccountId == -1 {
				imbalanced_account, err := GetImbalanceAccount(user.UserId, split.SecurityId)
				if err != nil {
					WriteError(w, 999 /*Internal Error*/)
					log.Print(err)
					return
				}

				split.AccountId = imbalanced_account.AccountId
				split.SecurityId = -1
			}
		}

		balanced, err := transaction.Balanced()
		if !balanced || err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}

		transactions = append(transactions, transaction)
	}

	for _, transaction := range transactions {
		err := InsertTransaction(&transaction, user)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
		}
	}

	WriteSuccess(w)
}
