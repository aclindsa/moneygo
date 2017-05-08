package main

import (
	"io"
	"log"
	"math/big"
	"net/http"
)

/*
 * Assumes the User is a valid, signed-in user, but accountid has not yet been validated
 */
func AccountImportHandler(w http.ResponseWriter, r *http.Request, user *User, accountid int64, importtype string) {
	//TODO branch off for different importtype's

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

	itl, err := ImportOFX(part)

	if err != nil {
		//TODO is this necessarily an invalid request (what if it was an error on our end)?
		WriteError(w, 3 /*Invalid Request*/)
		log.Print(err)
		return
	}

	if len(itl.Accounts) != 1 {
		WriteError(w, 3 /*Invalid Request*/)
		log.Printf("Found %d accounts when importing OFX, expected 1", len(itl.Accounts))
		return
	}

	sqltransaction, err := DB.Begin()
	if err != nil {
		WriteError(w, 999 /*Internal Error*/)
		log.Print(err)
		return
	}

	// Return Account with this Id
	account, err := GetAccountTx(sqltransaction, accountid, user.UserId)
	if err != nil {
		sqltransaction.Rollback()
		WriteError(w, 3 /*Invalid Request*/)
		log.Print(err)
		return
	}

	importedAccount := itl.Accounts[0]

	if len(account.ExternalAccountId) > 0 &&
		account.ExternalAccountId != importedAccount.ExternalAccountId {
		sqltransaction.Rollback()
		WriteError(w, 3 /*Invalid Request*/)
		log.Printf("OFX import has \"%s\" as ExternalAccountId, but the account being imported to has\"%s\"",
			importedAccount.ExternalAccountId,
			account.ExternalAccountId)
		return
	}

	if account.Type != importedAccount.Type {
		sqltransaction.Rollback()
		WriteError(w, 3 /*Invalid Request*/)
		log.Printf("Expected %s account, found %s in OFX file", account.Type.String(), importedAccount.Type.String())
		return
	}

	// Find matching existing securities or create new ones for those
	// referenced by the OFX import. Also create a map from placeholder import
	// SecurityIds to the actual SecurityIDs
	var securitymap = make(map[int64]*Security)
	for _, ofxsecurity := range itl.Securities {
		security, err := ImportGetCreateSecurity(sqltransaction, user, &ofxsecurity)
		if err != nil {
			sqltransaction.Rollback()
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
		securitymap[ofxsecurity.SecurityId] = security
	}

	if account.SecurityId != securitymap[importedAccount.SecurityId].SecurityId {
		sqltransaction.Rollback()
		WriteError(w, 3 /*Invalid Request*/)
		log.Printf("OFX import account's SecurityId (%d) does not match this account's (%d)", securitymap[importedAccount.SecurityId].SecurityId, account.SecurityId)
		return
	}

	// TODO Ensure all transactions have at least one split in the account
	// we're importing to?

	var transactions []Transaction
	for _, transaction := range itl.Transactions {
		transaction.UserId = user.UserId
		transaction.Status = Imported

		if !transaction.Valid() {
			sqltransaction.Rollback()
			WriteError(w, 999 /*Internal Error*/)
			log.Print("Unexpected invalid transaction from OFX import")
			return
		}

		// Ensure that either AccountId or SecurityId is set for this split,
		// and fixup the SecurityId to be a valid one for this user's actual
		// securities instead of a placeholder from the import
		for _, split := range transaction.Splits {
			if split.AccountId != -1 {
				if split.AccountId != importedAccount.AccountId {
					sqltransaction.Rollback()
					WriteError(w, 999 /*Internal Error*/)
					return
				}
				split.AccountId = account.AccountId
			} else if split.SecurityId != -1 {
				if sec, ok := securitymap[split.SecurityId]; ok {
					split.SecurityId = sec.SecurityId
				} else {
					sqltransaction.Rollback()
					WriteError(w, 999 /*Internal Error*/)
					log.Print("Couldn't find split's SecurityId in map during OFX import")
					return
				}
			} else {
				sqltransaction.Rollback()
				WriteError(w, 999 /*Internal Error*/)
				log.Print("Neither Split.AccountId Split.SecurityId was set during OFX import")
				return
			}
		}

		imbalances, err := transaction.GetImbalancesTx(sqltransaction)
		if err != nil {
			sqltransaction.Rollback()
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
					imbalanced_account, err = GetTradingAccount(sqltransaction, user.UserId, imbalanced_security)
				} else {
					imbalanced_account, err = GetImbalanceAccount(sqltransaction, user.UserId, imbalanced_security)
				}
				if err != nil {
					sqltransaction.Rollback()
					WriteError(w, 999 /*Internal Error*/)
					log.Print(err)
					return
				}

				// Add new split to fixup imbalance
				split := new(Split)
				r := new(big.Rat)
				r.Neg(&imbalance)
				security, err := GetSecurity(imbalanced_security, user.UserId)
				if err != nil {
					sqltransaction.Rollback()
					WriteError(w, 999 /*Internal Error*/)
					log.Print(err)
					return
				}
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
				imbalanced_account, err := GetImbalanceAccount(sqltransaction, user.UserId, split.SecurityId)
				if err != nil {
					sqltransaction.Rollback()
					WriteError(w, 999 /*Internal Error*/)
					log.Print(err)
					return
				}

				split.AccountId = imbalanced_account.AccountId
				split.SecurityId = -1
			}
		}

		transactions = append(transactions, transaction)
	}

	for _, transaction := range transactions {
		err := InsertTransactionTx(sqltransaction, &transaction, user)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
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
