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
                cash_flow = cash_flow - balance.amount
            end
        end
        series:value(month, cash_flow)
    end

    return t
end
