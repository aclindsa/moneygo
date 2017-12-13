package integration_test

import (
	"encoding/json"
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"net/http"
	"strings"
	"testing"
	"time"
)

// Needed because handlers.User doesn't allow Password to be written to JSON
type User struct {
	UserId          int64
	DefaultCurrency int64 // SecurityId of default currency, or ISO4217 code for it if creating new user
	Name            string
	Username        string
	Password        string
	PasswordHash    string
	Email           string
}

func (u *User) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(u)
}

func (u *User) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(u)
}

// TestData
type TestData struct {
	initialized  bool
	users        []User
	clients      []*http.Client
	securities   []models.Security
	prices       []models.Price
	accounts     []models.Account // accounts must appear after their parents in this slice
	transactions []models.Transaction
	reports      []models.Report
	tabulations  []models.Tabulation
}

type TestDataFunc func(*testing.T, *TestData)

func (t *TestData) initUser(user *User, userid int) error {
	newuser, err := createUser(user)
	if err != nil {
		return err
	}
	t.users = append(t.users, *newuser)

	// make a copy of the user so we can set the password for creating the
	// session without disturbing the original
	userWithPassword := *newuser
	userWithPassword.Password = user.Password

	client, err := newSession(&userWithPassword)
	if err != nil {
		return err
	}

	t.clients = append(t.clients, client)

	return nil
}

// Initialize makes requests to the server to create all of the objects
// represented in it before returning a copy of the data, with all of the *Id
// fields updated to their actual values
func (t *TestData) Initialize() (*TestData, error) {
	var t2 TestData
	for userid, user := range t.users {
		err := t2.initUser(&user, userid)
		if err != nil {
			return nil, err
		}
	}

	for _, security := range t.securities {
		s2, err := createSecurity(t2.clients[security.UserId], &security)
		if err != nil {
			return nil, err
		}
		t2.securities = append(t2.securities, *s2)
	}

	for _, price := range t.prices {
		userid := t.securities[price.SecurityId].UserId
		price.SecurityId = t2.securities[price.SecurityId].SecurityId
		price.CurrencyId = t2.securities[price.CurrencyId].SecurityId
		p2, err := createPrice(t2.clients[userid], &price)
		if err != nil {
			return nil, err
		}
		t2.prices = append(t2.prices, *p2)
	}

	for _, account := range t.accounts {
		account.SecurityId = t2.securities[account.SecurityId].SecurityId
		if account.ParentAccountId != -1 {
			account.ParentAccountId = t2.accounts[account.ParentAccountId].AccountId
		}
		a2, err := createAccount(t2.clients[account.UserId], &account)
		if err != nil {
			return nil, err
		}
		t2.accounts = append(t2.accounts, *a2)
	}

	for i, transaction := range t.transactions {
		transaction.Splits = []*models.Split{}
		for _, s := range t.transactions[i].Splits {
			// Make a copy of the split since Splits is a slice of pointers so
			// copying the transaction doesn't
			split := *s
			split.AccountId = t2.accounts[split.AccountId].AccountId
			transaction.Splits = append(transaction.Splits, &split)
		}
		tt2, err := createTransaction(t2.clients[transaction.UserId], &transaction)
		if err != nil {
			return nil, err
		}
		t2.transactions = append(t2.transactions, *tt2)
	}

	for _, report := range t.reports {
		r2, err := createReport(t2.clients[report.UserId], &report)
		if err != nil {
			return nil, err
		}
		t2.reports = append(t2.reports, *r2)
	}

	t2.initialized = true
	return &t2, nil
}

func (t *TestData) Teardown() error {
	if !t.initialized {
		return fmt.Errorf("Cannot teardown uninitialized TestData")
	}
	for userid, user := range t.users {
		err := deleteUser(t.clients[userid], &user)
		if err != nil {
			return err
		}
	}
	return nil
}

var data = []TestData{
	{
		users: []User{
			{
				DefaultCurrency: 840, // USD
				Name:            "John Smith",
				Username:        "jsmith",
				Password:        "hunter2",
				Email:           "jsmith@example.com",
			},
			{
				DefaultCurrency: 978, // Euro
				Name:            "Billy Bob",
				Username:        "bbob6",
				Password:        "#)$&!KF(*ADAHK#@*(FAJSDkalsdf98af32klhf98sd8a'2938LKJD",
				Email:           "bbob+moneygo@my-domain.com",
			},
		},
		securities: []models.Security{
			{
				UserId:      0,
				Name:        "USD",
				Description: "US Dollar",
				Symbol:      "$",
				Precision:   2,
				Type:        models.Currency,
				AlternateId: "840",
			},
			{
				UserId:      0,
				Name:        "SPY",
				Description: "SPDR S&P 500 ETF Trust",
				Symbol:      "SPY",
				Precision:   5,
				Type:        models.Stock,
				AlternateId: "78462F103",
			},
			{
				UserId:      1,
				Name:        "EUR",
				Description: "Euro",
				Symbol:      "€",
				Precision:   2,
				Type:        models.Currency,
				AlternateId: "978",
			},
			{
				UserId:      0,
				Name:        "EUR",
				Description: "Euro",
				Symbol:      "€",
				Precision:   2,
				Type:        models.Currency,
				AlternateId: "978",
			},
		},
		prices: []models.Price{
			{
				SecurityId: 1,
				CurrencyId: 0,
				Date:       time.Date(2017, time.January, 2, 21, 0, 0, 0, time.UTC),
				Value:      NewAmount("225.24"),
				RemoteId:   "12387-129831-1238",
			},
			{
				SecurityId: 1,
				CurrencyId: 0,
				Date:       time.Date(2017, time.January, 3, 21, 0, 0, 0, time.UTC),
				Value:      NewAmount("226.58"),
				RemoteId:   "12387-129831-1239",
			},
			{
				SecurityId: 1,
				CurrencyId: 0,
				Date:       time.Date(2017, time.January, 4, 21, 0, 0, 0, time.UTC),
				Value:      NewAmount("226.40"),
				RemoteId:   "12387-129831-1240",
			},
			{
				SecurityId: 1,
				CurrencyId: 0,
				Date:       time.Date(2017, time.January, 5, 21, 0, 0, 0, time.UTC),
				Value:      NewAmount("227.21"),
				RemoteId:   "12387-129831-1241",
			},
			{
				SecurityId: 0,
				CurrencyId: 3,
				Date:       time.Date(2017, time.November, 16, 18, 49, 53, 0, time.UTC),
				Value:      NewAmount("0.85"),
				RemoteId:   "USDEUR819298714",
			},
		},
		accounts: []models.Account{
			{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: -1,
				Type:            models.Asset,
				Name:            "Assets",
			},
			{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: 0,
				Type:            models.Bank,
				Name:            "Credit Union Checking",
			},
			{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: -1,
				Type:            models.Expense,
				Name:            "Expenses",
			},
			{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: 2,
				Type:            models.Expense,
				Name:            "Groceries",
			},
			{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: 2,
				Type:            models.Expense,
				Name:            "Cable",
			},
			{
				UserId:          1,
				SecurityId:      2,
				ParentAccountId: -1,
				Type:            models.Asset,
				Name:            "Assets",
			},
			{
				UserId:          1,
				SecurityId:      2,
				ParentAccountId: -1,
				Type:            models.Expense,
				Name:            "Expenses",
			},
			{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: -1,
				Type:            models.Liability,
				Name:            "Credit Card",
			},
		},
		transactions: []models.Transaction{
			{
				UserId:      0,
				Description: "weekly groceries",
				Date:        time.Date(2017, time.October, 15, 1, 16, 59, 0, time.UTC),
				Splits: []*models.Split{
					{
						Status:     models.Reconciled,
						AccountId:  1,
						SecurityId: -1,
						Amount:     NewAmount("-5.6"),
					},
					{
						Status:     models.Reconciled,
						AccountId:  3,
						SecurityId: -1,
						Amount:     NewAmount("5.6"),
					},
				},
			},
			{
				UserId:      0,
				Description: "weekly groceries",
				Date:        time.Date(2017, time.October, 31, 19, 10, 14, 0, time.UTC),
				Splits: []*models.Split{
					{
						Status:     models.Reconciled,
						AccountId:  1,
						SecurityId: -1,
						Amount:     NewAmount("-81.59"),
					},
					{
						Status:     models.Reconciled,
						AccountId:  3,
						SecurityId: -1,
						Amount:     NewAmount("81.59"),
					},
				},
			},
			{
				UserId:      0,
				Description: "Cable",
				Date:        time.Date(2017, time.September, 2, 0, 00, 00, 0, time.UTC),
				Splits: []*models.Split{
					{
						Status:     models.Reconciled,
						AccountId:  1,
						SecurityId: -1,
						Amount:     NewAmount("-39.99"),
					},
					{
						Status:     models.Entered,
						AccountId:  4,
						SecurityId: -1,
						Amount:     NewAmount("39.99"),
					},
				},
			},
			{
				UserId:      1,
				Description: "Gas",
				Date:        time.Date(2017, time.November, 1, 13, 19, 50, 0, time.UTC),
				Splits: []*models.Split{
					{
						Status:     models.Reconciled,
						AccountId:  5,
						SecurityId: -1,
						Amount:     NewAmount("-24.56"),
					},
					{
						Status:     models.Entered,
						AccountId:  6,
						SecurityId: -1,
						Amount:     NewAmount("24.56"),
					},
				},
			},
		},
		reports: []models.Report{
			{
				UserId: 0,
				Name:   "This Year's Monthly Expenses",
				Lua: `
function account_series_map(accounts, tabulation)
    map = {}

    for i=1,100 do -- we're not messing with accounts more than 100 levels deep
        all_handled = true
        for id, acct in pairs(accounts) do
            if not map[id] then
                all_handled = false
                if not acct.parent then
                    map[id] = tabulation:series(acct.name)
                elseif map[acct.parent.accountid] then
                    map[id] = map[acct.parent.accountid]:series(acct.name)
                end
            end
        end
        if all_handled then
            return map
        end
    end

    error("Accounts nested (at least) 100 levels deep")
end

function generate()
    year = 2017
    account_type = account.Expense

    accounts = get_accounts()
    t = tabulation.new(12)
    t:title(year .. " Monthly Expenses")
    t:subtitle("This is my subtitle")
    t:units(get_default_currency().Symbol)
    series_map = account_series_map(accounts, t)

    for month=1,12 do
        begin_date = date.new(year, month, 1)
        end_date = date.new(year, month+1, 1)

        t:label(month, tostring(begin_date))

        for id, acct in pairs(accounts) do
            series = series_map[id]
            if acct.type == account_type then
                balance = acct:balance(begin_date, end_date)
                series:value(month, balance.amount)
            end
        end
    end

    return t
end`,
			},
		},
		tabulations: []models.Tabulation{
			{
				ReportId: 0,
				Title:    "2017 Monthly Expenses",
				Subtitle: "This is my subtitle",
				Units:    "USD",
				Labels:   []string{"2017-01-01", "2017-02-01", "2017-03-01", "2017-04-01", "2017-05-01", "2017-06-01", "2017-07-01", "2017-08-01", "2017-09-01", "2017-10-01", "2017-11-01", "2017-12-01"},
				Series: map[string]*models.Series{
					"Assets": {
						Values: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						Series: map[string]*models.Series{
							"Credit Union Checking": {
								Values: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
								Series: map[string]*models.Series{},
							},
						},
					},
					"Expenses": {
						Values: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						Series: map[string]*models.Series{
							"Groceries": {
								Values: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 87.19, 0, 0},
								Series: map[string]*models.Series{},
							},
							"Cable": {
								Values: []float64{0, 0, 0, 0, 0, 0, 0, 0, 39.99, 0, 0, 0},
								Series: map[string]*models.Series{},
							},
						},
					},
					"Credit Card": {
						Values: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						Series: map[string]*models.Series{},
					},
				},
			},
		},
	},
}
