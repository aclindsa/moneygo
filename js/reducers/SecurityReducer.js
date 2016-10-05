var assign = require('object-assign');

var SecurityConstants = require('../constants/SecurityConstants');
var UserConstants = require('../constants/UserConstants');

module.exports = function(state = {}, action) {
	switch (action.type) {
		case SecurityConstants.SECURITIES_FETCHED:
			var securities = {};
			for (var i = 0; i < action.securities.length; i++) {
				var security = action.securities[i];
				securities[security.SecurityId] = security;
			}
			return securities;
		case SecurityConstants.SECURITY_CREATED:
		case SecurityConstants.SECURITY_UPDATED:
			var security = action.security;
			return assign({}, state, {
				[security.SecurityId]: security
			});
		case SecurityConstants.SECURITY_REMOVED:
			var newstate = assign({}, state);
			delete newstate[action.securityId];
			return newstate;
		case UserConstants.USER_LOGGEDOUT:
			return {};
		default:
			return state;
	}
};
