package handlers

import (
	"errors"
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"log"
	"net/http"
)

type UserExistsError struct{}

func (ueu UserExistsError) Error() string {
	return "User exists"
}

func GetUser(tx *Tx, userid int64) (*models.User, error) {
	var u models.User

	err := tx.SelectOne(&u, "SELECT * from users where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByUsername(tx *Tx, username string) (*models.User, error) {
	var u models.User

	err := tx.SelectOne(&u, "SELECT * from users where Username=?", username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func InsertUser(tx *Tx, u *models.User) error {
	security_template := FindCurrencyTemplate(u.DefaultCurrency)
	if security_template == nil {
		return errors.New("Invalid ISO4217 Default Currency")
	}

	existing, err := tx.SelectInt("SELECT count(*) from users where Username=?", u.Username)
	if err != nil {
		return err
	}
	if existing > 0 {
		return UserExistsError{}
	}

	err = tx.Insert(u)
	if err != nil {
		return err
	}

	// Copy the security template and give it our new UserId
	var security models.Security
	security = *security_template
	security.UserId = u.UserId

	err = InsertSecurity(tx, &security)
	if err != nil {
		return err
	}

	// Update the user's DefaultCurrency to our new SecurityId
	u.DefaultCurrency = security.SecurityId
	count, err := tx.Update(u)
	if err != nil {
		return err
	} else if count != 1 {
		return errors.New("Would have updated more than one user")
	}

	return nil
}

func GetUserFromSession(tx *Tx, r *http.Request) (*models.User, error) {
	s, err := GetSession(tx, r)
	if err != nil {
		return nil, err
	}
	return GetUser(tx, s.UserId)
}

func UpdateUser(tx *Tx, u *models.User) error {
	security, err := GetSecurity(tx, u.DefaultCurrency, u.UserId)
	if err != nil {
		return err
	} else if security.UserId != u.UserId || security.SecurityId != u.DefaultCurrency {
		return errors.New("UserId and DefaultCurrency don't match the fetched security")
	} else if security.Type != models.Currency {
		return errors.New("New DefaultCurrency security is not a currency")
	}

	count, err := tx.Update(u)
	if err != nil {
		return err
	} else if count != 1 {
		return errors.New("Would have updated more than one user")
	}

	return nil
}

func DeleteUser(tx *Tx, u *models.User) error {
	count, err := tx.Delete(u)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("No user to delete")
	}
	_, err = tx.Exec("DELETE FROM prices WHERE prices.SecurityId IN (SELECT securities.SecurityId FROM securities WHERE securities.UserId=?)", u.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM splits WHERE splits.TransactionId IN (SELECT transactions.TransactionId FROM transactions WHERE transactions.UserId=?)", u.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM transactions WHERE transactions.UserId=?", u.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM securities WHERE securities.UserId=?", u.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM accounts WHERE accounts.UserId=?", u.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM reports WHERE reports.UserId=?", u.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM sessions WHERE sessions.UserId=?", u.UserId)
	if err != nil {
		return err
	}

	return nil
}

func UserHandler(r *http.Request, context *Context) ResponseWriterWriter {
	if r.Method == "POST" {
		var user models.User
		if err := ReadJSON(r, &user); err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		user.UserId = -1
		user.HashPassword()

		err := InsertUser(context.Tx, &user)
		if err != nil {
			if _, ok := err.(UserExistsError); ok {
				return NewError(4 /*User Exists*/)
			} else {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
		}

		return ResponseWrapper{201, &user}
	} else {
		user, err := GetUserFromSession(context.Tx, r)
		if err != nil {
			return NewError(1 /*Not Signed In*/)
		}

		userid, err := context.NextID()
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}

		if userid != user.UserId {
			return NewError(2 /*Unauthorized Access*/)
		}

		if r.Method == "GET" {
			return user
		} else if r.Method == "PUT" {
			// Save old PWHash in case the new password is bogus
			old_pwhash := user.PasswordHash

			if err := ReadJSON(r, &user); err != nil || user.UserId != userid {
				return NewError(3 /*Invalid Request*/)
			}

			// If the user didn't create a new password, keep their old one
			if user.Password != models.BogusPassword {
				user.HashPassword()
			} else {
				user.Password = ""
				user.PasswordHash = old_pwhash
			}

			err = UpdateUser(context.Tx, user)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return user
		} else if r.Method == "DELETE" {
			err := DeleteUser(context.Tx, user)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}
