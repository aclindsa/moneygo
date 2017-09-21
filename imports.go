package main

import (
	"encoding/json"
	"github.com/aclindsa/ofxgo"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type OFXDownload struct {
	OFXPassword string
	StartDate   time.Time
	EndDate     time.Time
}

func (od *OFXDownload) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(od)
}

func ofxImportHelper(r io.Reader, w http.ResponseWriter, user *User, accountid int64) {
	itl, err := ImportOFX(r)

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

	// Find matching existing securities or create new ones for those
	// referenced by the OFX import. Also create a map from placeholder import
	// SecurityIds to the actual SecurityIDs
	var securitymap = make(map[int64]Security)
	for _, ofxsecurity := range itl.Securities {
		// save off since ImportGetCreateSecurity overwrites SecurityId on
		// ofxsecurity
		oldsecurityid := ofxsecurity.SecurityId
		security, err := ImportGetCreateSecurity(sqltransaction, user.UserId, &ofxsecurity)
		if err != nil {
			sqltransaction.Rollback()
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
		securitymap[oldsecurityid] = *security
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
			split.Status = Imported
			if split.AccountId != -1 {
				if split.AccountId != importedAccount.AccountId {
					sqltransaction.Rollback()
					WriteError(w, 999 /*Internal Error*/)
					log.Print("Imported split's AccountId wasn't -1 but also didn't match the account")
					return
				}
				split.AccountId = account.AccountId
			} else if split.SecurityId != -1 {
				if sec, ok := securitymap[split.SecurityId]; ok {
					// TODO try to auto-match splits to existing accounts based on past transactions that look like this one
					if split.ImportSplitType == TradingAccount {
						// Find/make trading account if we're that type of split
						trading_account, err := GetTradingAccount(sqltransaction, user.UserId, sec.SecurityId)
						if err != nil {
							sqltransaction.Rollback()
							WriteError(w, 999 /*Internal Error*/)
							log.Print("Couldn't find split's SecurityId in map during OFX import")
							return
						}
						split.AccountId = trading_account.AccountId
						split.SecurityId = -1
					} else {
						split.SecurityId = sec.SecurityId
					}
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
		for imbalanced_security, imbalance := range imbalances {
			if imbalance.Cmp(&zero) != 0 {
				imbalanced_account, err := GetImbalanceAccount(sqltransaction, user.UserId, imbalanced_security)
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
				security, err := GetSecurityTx(sqltransaction, imbalanced_security, user.UserId)
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
		// accounts. In the same loop, check to see if this transaction/split
		// has been imported before
		var already_imported bool
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

			exists, err := split.AlreadyImportedTx(sqltransaction)
			if err != nil {
				sqltransaction.Rollback()
				WriteError(w, 999 /*Internal Error*/)
				log.Print("Error checking if split was already imported:", err)
				return
			} else if exists {
				already_imported = true
			}
		}

		if !already_imported {
			transactions = append(transactions, transaction)
		}
	}

	for _, transaction := range transactions {
		err := InsertTransactionTx(sqltransaction, &transaction, user)
		if err != nil {
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

func OFXImportHandler(w http.ResponseWriter, r *http.Request, user *User, accountid int64) {
	download_json := r.PostFormValue("ofxdownload")
	if download_json == "" {
		log.Print("download_json")
		WriteError(w, 3 /*Invalid Request*/)
		return
	}

	var ofxdownload OFXDownload
	err := ofxdownload.Read(download_json)
	if err != nil {
		log.Print("ofxdownload.Read")
		WriteError(w, 3 /*Invalid Request*/)
		return
	}

	account, err := GetAccount(accountid, user.UserId)
	if err != nil {
		log.Print("GetAccount")
		WriteError(w, 3 /*Invalid Request*/)
		return
	}

	ofxver := ofxgo.OfxVersion203
	if len(account.OFXVersion) != 0 {
		ofxver, err = ofxgo.NewOfxVersion(account.OFXVersion)
		if err != nil {
			log.Print("NewOfxVersion")
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
	}

	var client = ofxgo.Client{
		AppID:       account.OFXAppID,
		AppVer:      account.OFXAppVer,
		SpecVersion: ofxver,
		NoIndent:    account.OFXNoIndent,
	}

	var query ofxgo.Request
	query.URL = account.OFXURL
	query.Signon.ClientUID = ofxgo.UID(account.OFXClientUID)
	query.Signon.UserID = ofxgo.String(account.OFXUser)
	query.Signon.UserPass = ofxgo.String(ofxdownload.OFXPassword)
	query.Signon.Org = ofxgo.String(account.OFXORG)
	query.Signon.Fid = ofxgo.String(account.OFXFID)

	transactionuid, err := ofxgo.RandomUID()
	if err != nil {
		WriteError(w, 999 /*Internal Error*/)
		log.Println("Error creating uid for transaction:", err)
		return
	}

	if account.Type == Investment {
		// Investment account
		statementRequest := ofxgo.InvStatementRequest{
			TrnUID: *transactionuid,
			InvAcctFrom: ofxgo.InvAcct{
				BrokerID: ofxgo.String(account.OFXBankID),
				AcctID:   ofxgo.String(account.OFXAcctID),
			},
			Include:        true,
			IncludeOO:      true,
			IncludePos:     true,
			IncludeBalance: true,
			Include401K:    true,
			Include401KBal: true,
		}
		query.InvStmt = append(query.InvStmt, &statementRequest)
	} else if account.OFXAcctType == "CC" {
		// Import credit card transactions
		statementRequest := ofxgo.CCStatementRequest{
			TrnUID: *transactionuid,
			CCAcctFrom: ofxgo.CCAcct{
				AcctID: ofxgo.String(account.OFXAcctID),
			},
			Include: true,
		}
		query.CreditCard = append(query.CreditCard, &statementRequest)
	} else {
		// Import generic bank transactions
		acctTypeEnum, err := ofxgo.NewAcctType(account.OFXAcctType)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		statementRequest := ofxgo.StatementRequest{
			TrnUID: *transactionuid,
			BankAcctFrom: ofxgo.BankAcct{
				BankID:   ofxgo.String(account.OFXBankID),
				AcctID:   ofxgo.String(account.OFXAcctID),
				AcctType: acctTypeEnum,
			},
			Include: true,
		}
		query.Bank = append(query.Bank, &statementRequest)
	}

	response, err := client.RequestNoParse(&query)
	if err != nil {
		// TODO this could be an error talking with the OFX server...
		WriteError(w, 3 /*Invalid Request*/)
		log.Print(err)
		return
	}
	defer response.Body.Close()

	ofxImportHelper(response.Body, w, user, accountid)
}

func OFXFileImportHandler(w http.ResponseWriter, r *http.Request, user *User, accountid int64) {
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
			log.Print("Encountered unexpected EOF")
		} else {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
		}
		return
	}

	ofxImportHelper(part, w, user, accountid)
}

/*
 * Assumes the User is a valid, signed-in user, but accountid has not yet been validated
 */
func AccountImportHandler(w http.ResponseWriter, r *http.Request, user *User, accountid int64, importtype string) {

	switch importtype {
	case "ofx":
		OFXImportHandler(w, r, user, accountid)
	case "ofxfile":
		OFXFileImportHandler(w, r, user, accountid)
	default:
		WriteError(w, 3 /*Invalid Request*/)
	}
}
