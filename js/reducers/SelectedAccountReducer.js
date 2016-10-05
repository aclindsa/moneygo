var AccountConstants = require('../constants/AccountConstants');
var UserConstants = require('../constants/UserConstants');

module.exports = function(state = -1, action) {
	switch (action.type) {
		case AccountConstants.ACCOUNTS_FETCHED:
			for (var i = 0; i < action.accounts.length; i++) {
				if (action.accounts[i].AccountId == state)
					return state;
			}
			return -1;
		case AccountConstants.ACCOUNT_REMOVED:
			if (action.accountId == state)
				return -1;
			return state;
		case AccountConstants.ACCOUNT_SELECTED:
			return action.accountId;
		case UserConstants.USER_LOGGEDOUT:
			return -1;
		default:
			return state;
	}
};
