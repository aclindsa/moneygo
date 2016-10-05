var assign = require('object-assign');

var AccountConstants = require('../constants/AccountConstants');
var UserConstants = require('../constants/UserConstants');

function accountChildren(accounts) {
	var children = {};
	for (var accountId in accounts) {
		if (accounts.hasOwnProperty(accountId)) {
			var parentAccountId = accounts[accountId].ParentAccountId;
			if (!children.hasOwnProperty(parentAccountId))
				children[parentAccountId] = [];
			if (!children.hasOwnProperty(accountId))
				children[accountId] = [];
			children[parentAccountId].push(accountId);
		}
	}
	return children;
}

const initialState = {map: {}, children: {}};

module.exports = function(state = initialState, action) {
	switch (action.type) {
		case AccountConstants.ACCOUNTS_FETCHED:
			var accounts = {};
			for (var i = 0; i < action.accounts.length; i++) {
				var account = action.accounts[i];
				accounts[account.AccountId] = account;
			}
			return {
				map: accounts,
				children: accountChildren(accounts)
			};
		case AccountConstants.ACCOUNT_CREATED:
		case AccountConstants.ACCOUNT_UPDATED:
			var account = action.account;
			var accounts = assign({}, state.map, {
				[account.AccountId]: account
			});
			return {
				map: accounts,
				children: accountChildren(accounts)
			};
		case AccountConstants.ACCOUNT_REMOVED:
			var accounts = assign({}, state.map);
			delete accounts[action.accountId];
			return {
				map: accounts,
				children: accountChildren(accounts)
			};
		case UserConstants.USER_LOGGEDOUT:
			return initialState;
		default:
			return state;
	}
};
