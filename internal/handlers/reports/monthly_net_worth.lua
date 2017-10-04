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

    accounts = get_accounts()
    t = tabulation.new(12)
    t:title(year .. " Monthly Net Worth")
    series_map = account_series_map(accounts, t)
    default_currency = get_default_currency()

    for month=1,12 do
        end_date = date.new(year, month+1, 1)

        t:label(month, tostring(end_date))

        for id, acct in pairs(accounts) do
            series = series_map[id]
            if acct.type ~= account.Expense and acct.type ~= account.Income and acct.type ~= account.Trading then
                balance = acct:balance(end_date)
                multiplier = 1
                if acct.security ~= default_currency and balance.amount ~= 0 then
                    price = acct.security:closestprice(default_currency, end_date)
                    if price == nil then
                        --[[
                        -- This should contain code to warn the user that their report is missing some information
                        --]]
                        multiplier = 0
                    else
                        multiplier = price.value
                    end
                end
                series:value(month, balance.amount * multiplier)
            end
        end
    end

    return t
end
