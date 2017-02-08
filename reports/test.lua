accounts = get_accounts()

for id, account in pairs(accounts) do
	print(account, account.security)
	a = account:balance(date.new("2015-12-01"), date.new("2017-12-01"))
	b = account:balance(date.new("2015-06-01"), date.new("2015-12-01"))
	print(a, b, a+b, account:balance())
end
