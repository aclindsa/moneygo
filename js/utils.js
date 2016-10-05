const recursiveAccountDisplayInfo = function(account, account_map, accountChildren, prefix) {
	var name = prefix + account.Name;
	var accounts = [{AccountId: account.AccountId, Name: name}];
	for (var i = 0; i < accountChildren[account.AccountId].length; i++)
		accounts = accounts.concat(recursiveAccountDisplayInfo(account_map[accountChildren[account.AccountId][i]], account_map, accountChildren, name + "/"));
	return accounts
};

const getAccountDisplayList = function(account_map, accountChildren, includeRoot, rootName) {
	var accounts = []
	if (includeRoot)
		accounts.push({AccountId: -1, Name: rootName});
	for (var accountId in account_map) {
		if (account_map.hasOwnProperty(accountId) &&
				account_map[accountId].isRootAccount())
			accounts = accounts.concat(recursiveAccountDisplayInfo(account_map[accountId], account_map, accountChildren, ""));
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
