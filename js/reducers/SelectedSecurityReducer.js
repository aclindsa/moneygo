var SecurityConstants = require('../constants/SecurityConstants');
var UserConstants = require('../constants/UserConstants');

module.exports = function(state = -1, action) {
	switch (action.type) {
		case SecurityConstants.SECURITIES_FETCHED:
			for (var i = 0; i < action.securities.length; i++) {
				if (action.securities[i].SecurityId == state)
					return state;
			}
			return -1;
		case SecurityConstants.SECURITY_REMOVED:
			if (action.securityId == state)
				return -1;
			return state;
		case SecurityConstants.SECURITY_SELECTED:
			return action.securityId;
		case UserConstants.USER_LOGGEDOUT:
			return -1;
		default:
			return state;
	}
};
