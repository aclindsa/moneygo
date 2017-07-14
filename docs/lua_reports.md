# Lua Reports

MoneyGo reports are written in [Lua](https://lua.org), as implemented by
[github.com/yuin/gopher-lua](https://github.com/yuin/gopher-lua), with hooks
added to query the necessary MoneyGo state to generate the report.

## An Example: Monthly Cash Flow Report

Before diving into the details, here's an example report that calculates the
difference between income and expenses for each month in the current year:

```
function generate()
  year = date.now().year

  accounts = get_accounts()
  t = tabulation.new(12)
  t:title(year .. " Monthly Cash Flow")
  series = t:series("Income minus expenses")

  for month=1,12 do
    begin_date = date.new(year, month, 1)
    end_date = date.new(year, month+1, 1)

    t:label(month, tostring(begin_date))
    cash_flow = 0

    for id, acct in pairs(accounts) do
      if acct.type == account.Expense or acct.type == account.Income then
        balance = acct:balance(begin_date, end_date)
        --[[
        Note: We should convert balance.amount to the user's default currency
        before proceeding here
        --]]
        cash_flow = cash_flow - balance.amount
      end
    end
    series:value(month, cash_flow)
  end

  return t
end
```

More examples can be found in the reports/ directory in the MoneyGo source tree.

## Basic Operation

The lua code behind a report *must* contain a `generate()` function which takes
no arguments. This function is called when generating a report, and must return
a `tabulation` object, created by calling `t = tabulation:new(n)`, where `n` is
the integer number of data values in each of the series of this tabulation (all
series in the same tabulation must have the same number of values).

### Titles and Labels

Assuming your tabulation object is `t`, you should then call `t.label(m,
"some_string")` for each value of `m` in `[1, n]` to set the label for the 'm'th
data element in each series. You do not need to do this before creating series,
and can do it lazily as you generate data, if needed. Titles, subtitles, and the
y-axis label (the units) can be set as follows on a tabulation object named `t`:

* `t.title("The title of my report")`
* `t.subtitle("The subtitle of my report")`
* `t.units("USD ($)")`

### Data Series

To create a new top-level series, call `s = t:series("series name")` where `t`
is a tabulation object. Just as for labels for tabulation objects, you set
`s:value(m, number)` for each value of `m` in `[1, n]` (where `n` is the same
integer used to create the tabulation object to which this series belongs).

Nested series can be created by calling `s2 = s:series("nested series name")`,
where `s` is any already-created series object. Nested series allow for drilling
down to explore the information in more detail. In the web interface, they are
clickable and cause the charts to display the selected series as the new top
level.

It is assumed that nested series' reported values are not already included in
their parents' values. When being displayed, all the children series values are
added into the parent's. This means that a series may have no values of its own,
but still show up in a chart because it is reporting the sum of its children's
values.

## Gathering Data

Collecting/tabulating the data is up to you (chasing this flexibility was the
impetus behind Lua reports in the first place).

### Accounts and Balances

You can get a table of account objects for each of your accounts (indexed by
account ID) by calling the global function, `get_accounts()`. Each of these
accounts has several fields describing it:

* `a.Name` returns the account's name
* `a.Description` returns the account's description
* `a.Type` returns the account's type, as an integer constant. The account type
  constants are available on the top-level 'account' object
   * `account.Bank`
   * `account.Cash`
   * `account.Asset`
   * `account.Liability`
   * `account.Investment`
   * `account.Income`
   * `account.Expense`
   * `account.Trading`
   * `account.Equity`
   * `account.Receivable`
   * `account.Payable`
* `a.TypeName` returns a string representation of the account's type
* `a.Security` returns a security object representing the currency, stock, etc.
  of this account.
* `a.Parent` returns this account's parent account, or nil if the account has no
  parent.
* `a:Balance` is a function which returns the account balance in the account's
  security, optionally over a date range. If no arguments are provided, the
  total account balance as of the end of time is returned. If one date is
  provided, the balance as of that date is returned. If two dates are provided,
  the difference in balances between the first and second dates is returned.

### Securities

You can get a table containing all the securities/currencies registered to an
account using the global function `get_securities()`. Each of these securities
has several fields describing it:

* `s.SecurityId`
* `s.Name`
* `s.Description`
* `s.Symbol` returns the symbol traditionally associated with that security
  (i.e. '$' for USD, or BRK.B for class B shares of Berkshire Hathaway)
* `s.Precision` returns the number of digits of precision past the decimal point
  that this currency allows for (i.e. 2 for USD)
* `s.Type` returns an int constant which represents what type of security it is
  (i.e. stock or currency)

Securities support a ClosestPrice function that allows you to fetch the price of
the current security in a given currency that is closest to the supplied date.
For example, to print the price in the user's default currency for each security
in the user's account:

```
  default_currency = get_default_currency()
  for id, security in pairs(get_securities()) do
    price = security.price(default_currency, date.now())
    if price ~= nil then
      print(tostring(security) .. ": " security.Symbol .. " " .. price.Value)
    else
      print("Failed to fetch price for " .. tostring(security))
    end
  end
```

You can also query for an account's default currency using the global
`get_default_currency()` function.

### Prices

Price objects can be queried from Security objects. Price objects contain the
following fields:

* `p.PriceId`
* `p.Security` returns the security object the price is for
* `p.Currency` returns the currency that the price is in
* `p.Value` returns the price of one unit of 'security' in 'currency', as a
  float

### Dates

In order to make it easier to do operations like finding account balances for a
month at a time, MoneyGo implements it's own date type (eschewing the
traditional Lua implementation). You *must* use the MoneyGo date types when
passing them to any MoneyGo lua functions. To create a date object, you can use
one of two methods:

1. `date.now()` returns the current date
2. `date.new(2017, 7, 5)` returns a date object representing July 5th, 2017.
   Note that this method also accepts a single argument of a table with the
   'year', 'month', and 'day' fields set to int's.

In addition to supporting conversion to a string, addition, subtraction, and
comparison operators, dates support returning their constituent parts using
`d.Year`, `d.Month`, and `d.Day`.
