package handlers

import (
	"errors"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
	"log"
	"net/http"
)

type UserExistsError struct{}

func (ueu UserExistsError) Error() string {
	return "User exists"
}

func InsertUser(tx store.Tx, u *models.User) error {
	security_template := FindCurrencyTemplate(u.DefaultCurrency)
	if security_template == nil {
		return errors.New("Invalid ISO4217 Default Currency")
	}

	exists, err := tx.UsernameExists(u.Username)
	if err != nil {
		return err
	}
	if exists {
		return UserExistsError{}
	}

	err = tx.InsertUser(u)
	if err != nil {
		return err
	}

	// Copy the security template and give it our new UserId
	var security models.Security
	security = *security_template
	security.UserId = u.UserId

	err = tx.InsertSecurity(&security)
	if err != nil {
		return err
	}

	// Update the user's DefaultCurrency to our new SecurityId
	u.DefaultCurrency = security.SecurityId
	err = tx.UpdateUser(u)
	if err != nil {
		return err
	}

	return nil
}

func GetUserFromSession(tx store.Tx, r *http.Request) (*models.User, error) {
	s, err := GetSession(tx, r)
	if err != nil {
		return nil, err
	}
	return tx.GetUser(s.UserId)
}

func UpdateUser(tx store.Tx, u *models.User) error {
	security, err := tx.GetSecurity(u.DefaultCurrency, u.UserId)
	if err != nil {
		return err
	} else if security.UserId != u.UserId || security.SecurityId != u.DefaultCurrency {
		return errors.New("UserId and DefaultCurrency don't match the fetched security")
	} else if security.Type != models.Currency {
		return errors.New("New DefaultCurrency security is not a currency")
	}

	err = tx.UpdateUser(u)
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
			err := context.Tx.DeleteUser(user)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}
