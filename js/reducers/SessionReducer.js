var UserConstants = require('../constants/UserConstants');

var Session = require('../models').Session;

module.exports = function(state = new Session(), action) {
	switch (action.type) {
		case UserConstants.USER_LOGGEDIN:
			return action.session;
		case UserConstants.USER_LOGGEDOUT:
			return new Session();
		default:
			return state;
	}
};
