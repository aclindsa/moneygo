package main

import (
	"io"
	"io/ioutil"
	"log"
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

	for _, transaction := range *itl.Transactions {
		if !transaction.Valid() {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		// TODO check if transactions are balanced too
		//		balanced, err := transaction.Balanced()
		//		if !balanced || err != nil {
		//			WriteError(w, 3 /*Invalid Request*/)
		//			return
		//		}
	}

	/////////////////////// TODO ////////////////////////
	for _, transaction := range *itl.Transactions {
		transaction.UserId = user.UserId
		transaction.Status = Imported
		err := InsertTransaction(&transaction, user)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
		}
	}

	WriteSuccess(w)
}
