package handlers

import (
	"encoding/json"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
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

func ofxImportHelper(tx store.Tx, r io.Reader, user *models.User, accountid int64) ResponseWriterWriter {
	itl, err := ImportOFX(r)

	if err != nil {
		//TODO is this necessarily an invalid request (what if it was an error on our end)?
		log.Print(err)
		return NewError(3 /*Invalid Request*/)
	}

	if len(itl.Accounts) != 1 {
		log.Printf("Found %d accounts when importing OFX, expected 1", len(itl.Accounts))
		return NewError(3 /*Invalid Request*/)
	}

	// Return Account with this Id
	account, err := tx.GetAccount(accountid, user.UserId)
	if err != nil {
		log.Print(err)
		return NewError(3 /*Invalid Request*/)
	}

	importedAccount := itl.Accounts[0]

	if len(account.ExternalAccountId) > 0 &&
		account.ExternalAccountId != importedAccount.ExternalAccountId {
		log.Printf("OFX import has \"%s\" as ExternalAccountId, but the account being imported to has\"%s\"",
			importedAccount.ExternalAccountId,
			account.ExternalAccountId)
		return NewError(3 /*Invalid Request*/)
	}

	// Find matching existing securities or create new ones for those
	// referenced by the OFX import. Also create a map from placeholder import
	// SecurityIds to the actual SecurityIDs
	var securitymap = make(map[int64]models.Security)
	for _, ofxsecurity := range itl.Securities {
		// save off since ImportGetCreateSecurity overwrites SecurityId on
		// ofxsecurity
		oldsecurityid := ofxsecurity.SecurityId
		security, err := ImportGetCreateSecurity(tx, user.UserId, &ofxsecurity)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}
		securitymap[oldsecurityid] = *security
	}

	if account.SecurityId != securitymap[importedAccount.SecurityId].SecurityId {
		log.Printf("OFX import account's SecurityId (%d) does not match this account's (%d)", securitymap[importedAccount.SecurityId].SecurityId, account.SecurityId)
		return NewError(3 /*Invalid Request*/)
	}

	// TODO Ensure all transactions have at least one split in the account
	// we're importing to?

	var transactions []models.Transaction
	for _, transaction := range itl.Transactions {
		transaction.UserId = user.UserId

		if !transaction.Valid() {
			log.Print("Unexpected invalid transaction from OFX import")
			return NewError(999 /*Internal Error*/)
		}

		// Ensure that either AccountId or SecurityId is set for this split,
		// and fixup the SecurityId to be a valid one for this user's actual
		// securities instead of a placeholder from the import
		for _, split := range transaction.Splits {
			split.Status = models.Imported
			if split.AccountId != -1 {
				if split.AccountId != importedAccount.AccountId {
					log.Print("Imported split's AccountId wasn't -1 but also didn't match the account")
					return NewError(999 /*Internal Error*/)
				}
				split.AccountId = account.AccountId
			} else if split.SecurityId != -1 {
				if sec, ok := securitymap[split.SecurityId]; ok {
					// TODO try to auto-match splits to existing accounts based on past transactions that look like this one
					if split.ImportSplitType == models.TradingAccount {
						// Find/make trading account if we're that type of split
						trading_account, err := GetTradingAccount(tx, user.UserId, sec.SecurityId)
						if err != nil {
							log.Print("Couldn't find split's SecurityId in map during OFX import")
							return NewError(999 /*Internal Error*/)
						}
						split.AccountId = trading_account.AccountId
						split.SecurityId = -1
					} else if split.ImportSplitType == models.SubAccount {
						subaccount := &models.Account{
							UserId:          user.UserId,
							Name:            sec.Name,
							ParentAccountId: account.AccountId,
							SecurityId:      sec.SecurityId,
							Type:            account.Type,
						}
						subaccount, err := GetCreateAccount(tx, *subaccount)
						if err != nil {
							log.Print(err)
							return NewError(999 /*Internal Error*/)
						}
						split.AccountId = subaccount.AccountId
						split.SecurityId = -1
					} else {
						split.SecurityId = sec.SecurityId
					}
				} else {
					log.Print("Couldn't find split's SecurityId in map during OFX import")
					return NewError(999 /*Internal Error*/)
				}
			} else {
				log.Print("Neither Split.AccountId Split.SecurityId was set during OFX import")
				return NewError(999 /*Internal Error*/)
			}
		}

		imbalances, err := GetTransactionImbalances(tx, &transaction)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}

		// Fixup any imbalances in transactions
		var zero big.Rat
		for imbalanced_security, imbalance := range imbalances {
			if imbalance.Cmp(&zero) != 0 {
				imbalanced_account, err := GetImbalanceAccount(tx, user.UserId, imbalanced_security)
				if err != nil {
					log.Print(err)
					return NewError(999 /*Internal Error*/)
				}

				// Add new split to fixup imbalance
				split := new(models.Split)
				r := new(big.Rat)
				r.Neg(&imbalance)
				security, err := tx.GetSecurity(imbalanced_security, user.UserId)
				if err != nil {
					log.Print(err)
					return NewError(999 /*Internal Error*/)
				}
				split.Amount.Rat = *r
				if split.Amount.Precision() > security.Precision {
					log.Printf("Precision on created imbalance-correction split (%d) greater than the underlying security (%s) allows (%d)", split.Amount.Precision(), security, security.Precision)
					return NewError(999 /*Internal Error*/)
				}
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
				imbalanced_account, err := GetImbalanceAccount(tx, user.UserId, split.SecurityId)
				if err != nil {
					log.Print(err)
					return NewError(999 /*Internal Error*/)
				}

				split.AccountId = imbalanced_account.AccountId
				split.SecurityId = -1
			}

			exists, err := tx.SplitExists(split)
			if err != nil {
				log.Print("Error checking if split was already imported:", err)
				return NewError(999 /*Internal Error*/)
			} else if exists {
				already_imported = true
			}
		}

		if !already_imported {
			transactions = append(transactions, transaction)
		}
	}

	for _, transaction := range transactions {
		err := tx.InsertTransaction(&transaction, user)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}
	}

	return SuccessWriter{}
}

func OFXImportHandler(context *Context, r *http.Request, user *models.User, accountid int64) ResponseWriterWriter {
	var ofxdownload OFXDownload
	if err := ReadJSON(r, &ofxdownload); err != nil {
		return NewError(3 /*Invalid Request*/)
	}

	account, err := context.Tx.GetAccount(accountid, user.UserId)
	if err != nil {
		return NewError(3 /*Invalid Request*/)
	}

	ofxver := ofxgo.OfxVersion203
	if len(account.OFXVersion) != 0 {
		ofxver, err = ofxgo.NewOfxVersion(account.OFXVersion)
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}
	}

	var client = ofxgo.BasicClient{
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
		log.Println("Error creating uid for transaction:", err)
		return NewError(999 /*Internal Error*/)
	}

	if account.Type == models.Investment {
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
			return NewError(3 /*Invalid Request*/)
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
		log.Print(err)
		return NewError(3 /*Invalid Request*/)
	}
	defer response.Body.Close()

	return ofxImportHelper(context.Tx, response.Body, user, accountid)
}

func OFXFileImportHandler(context *Context, r *http.Request, user *models.User, accountid int64) ResponseWriterWriter {
	multipartReader, err := r.MultipartReader()
	if err != nil {
		return NewError(3 /*Invalid Request*/)
	}

	// assume there is only one 'part'
	part, err := multipartReader.NextPart()
	if err != nil {
		if err == io.EOF {
			log.Print("Encountered unexpected EOF")
			return NewError(3 /*Invalid Request*/)
		} else {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}
	}

	return ofxImportHelper(context.Tx, part, user, accountid)
}

/*
 * Assumes the User is a valid, signed-in user, but accountid has not yet been validated
 */
func AccountImportHandler(context *Context, r *http.Request, user *models.User, accountid int64) ResponseWriterWriter {

	importType := context.NextLevel()
	switch importType {
	case "ofx":
		return OFXImportHandler(context, r, user, accountid)
	case "ofxfile":
		return OFXFileImportHandler(context, r, user, accountid)
	default:
		return NewError(3 /*Invalid Request*/)
	}
}

func ImportHandler(r *http.Request, context *Context) ResponseWriterWriter {
	route := context.NextLevel()
	if route != "gnucash" {
		return NewError(3 /*Invalid Request*/)
	}
	return GnucashImportHandler(r, context)
}
