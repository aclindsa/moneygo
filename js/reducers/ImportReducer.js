var assign = require('object-assign');

var ImportConstants = require('../constants/ImportConstants');
var UserConstants = require('../constants/UserConstants');

const initialState = {
	showModal: false,
	importing: false,
	uploadProgress: 0,
	importFinished: false,
	importFailed: false,
	errorMessage: null
};

module.exports = function(state = initialState, action) {
	switch (action.type) {
		case ImportConstants.OPEN_IMPORT_MODAL:
			return assign({}, initialState, {
				showModal: true
			});
		case ImportConstants.CLOSE_IMPORT_MODAL:
		case UserConstants.USER_LOGGEDOUT:
			return initialState;
		case ImportConstants.BEGIN_IMPORT:
			return assign({}, state, {
				importing: true
			});
		case ImportConstants.UPDATE_IMPORT_PROGRESS:
			return assign({}, state, {
				uploadProgress: action.progress
			});
		case ImportConstants.IMPORT_FINISHED:
			return assign({}, state, {
				importing: false,
				uploadProgress: 100,
				importFinished: true
			});
		case ImportConstants.IMPORT_FAILED:
			return assign({}, state, {
				importing: false,
				importFailed: true,
				errorMessage: action.error
			});
		default:
			return state;
	}
};
