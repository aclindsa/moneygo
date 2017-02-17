var assign = require('object-assign');

var ReportConstants = require('../constants/ReportConstants');
var UserConstants = require('../constants/UserConstants');

const initialState = {
	report: null,
	seriesTraversal: []
};

module.exports = function(state = initialState, action) {
	switch (action.type) {
		case ReportConstants.REPORT_SELECTED:
			return {
				report: action.report,
				seriesTraversal: action.seriesTraversal
			};
		case UserConstants.USER_LOGGEDOUT:
			return initialState;
		default:
			return state;
	}
};
