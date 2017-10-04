function generate()
    accounts = get_accounts()
    securities = get_securities()
    default_currency = get_default_currency()
    series_map = {}
    totals_map = {}

    t = tabulation.new(1)
    t:title("Current Asset Allocation")

    t:label(1, "Assets")

    for id, security in pairs(securities) do
        totals_map[id] = 0
        series_map[id] = t:series(tostring(security))
    end

    for id, acct in pairs(accounts) do
        if acct.type == account.Asset or acct.type == account.Investment or acct.type == account.Bank or acct.type == account.Cash then
            balance = acct:balance()
            multiplier = 1
            if acct.security ~= default_currency and balance.amount ~= 0 then
                price = acct.security:closestprice(default_currency, date.now())
                if price == nil then
                    --[[
                    -- This should contain code to warn the user that their report is missing some information
                    --]]
                    multiplier = 0
                else
                    multiplier = price.value
                end
            end
            totals_map[acct.security.SecurityId] = balance.amount * multiplier + totals_map[acct.security.SecurityId]
        end
    end

    for id, series in pairs(series_map) do
        series:value(1, totals_map[id])
    end

    return t
end
