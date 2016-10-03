var assign = require('object-assign');

var AccountConstants = require('../constants/AccountConstants');

module.exports = function(state = {}, action) {
	switch (action.type) {
		case AccountConstants.ACCOUNTS_FETCHED:
			var accounts = {};
			for (var i = 0; i < action.accounts.length; i++) {
				var account = action.accounts[i];
				accounts[account.AccountId] = account;
			}
			return accounts;
		case AccountConstants.ACCOUNT_CREATED:
		case AccountConstants.ACCOUNT_UPDATED:
			var account = action.account;
			return assign({}, state, {
				[account.AccountId]: account
			});
		case AccountConstants.ACCOUNT_REMOVED:
			var newstate = assign({}, state);
			delete newstate[action.accountId];
			return newstate;
		default:
			return state;
	}
};
