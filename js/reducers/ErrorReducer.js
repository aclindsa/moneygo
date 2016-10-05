var ErrorConstants = require('../constants/ErrorConstants');

var Error = require('../models').Error;

module.exports = function(state = new Error(), action) {
	switch (action.type) {
		case ErrorConstants.ERROR_AJAX:
		case ErrorConstants.ERROR_SERVER:
		case ErrorConstants.ERROR_CLIENT:
		case ErrorConstants.ERROR_USER:
			return action.error;
		case ErrorConstants.CLEAR_ERROR:
			return new Error();
		default:
			return state;
	}
};
