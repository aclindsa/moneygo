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
    year = date.now().year
    account_type = account.Expense

    accounts = get_accounts()
    t = tabulation.new(12)
    t:title(year .. " Monthly Expenses")
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
end
