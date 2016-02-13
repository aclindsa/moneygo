const recursiveAccountDisplayInfo = function(account, prefix) {
	var name = prefix + account.Name;
	var accounts = [{AccountId: account.AccountId, Name: name}];
	for (var i = 0; i < account.Children.length; i++)
		accounts = accounts.concat(recursiveAccountDisplayInfo(account.Children[i], name + "/"));
	return accounts
};

const getAccountDisplayList = function(account_list, includeRoot, rootName) {
	var accounts = []
	if (includeRoot)
		accounts.push({AccountId: -1, Name: rootName});
	for (var i = 0; i < account_list.length; i++) {
		if (account_list[i].isRootAccount())
			accounts = accounts.concat(recursiveAccountDisplayInfo(account_list[i], ""));
	}
	return accounts;
};

const getAccountDisplayName = function(account, account_map) {
	var name = account.Name;
	while (account.ParentAccountId >= 0) {
		account = account_map[account.ParentAccountId];
		name = account.Name + "/" + name;
	}
	return name;
};

module.exports = {
	getAccountDisplayList: getAccountDisplayList,
	getAccountDisplayName: getAccountDisplayName
};
