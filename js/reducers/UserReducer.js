var UserConstants = require('../constants/UserConstants');

var User = require('../models').User;

module.exports = function(state = new User(), action) {
	switch (action.type) {
		case UserConstants.USER_FETCHED:
		case UserConstants.USER_UPDATED:
			return action.user;
		case UserConstants.USER_LOGGEDOUT:
			return new User();
		default:
			return state;
	}
};
