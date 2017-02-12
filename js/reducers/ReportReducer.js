var assign = require('object-assign');

var ReportConstants = require('../constants/ReportConstants');
var UserConstants = require('../constants/UserConstants');

module.exports = function(state = {}, action) {
	switch (action.type) {
		case ReportConstants.REPORT_FETCHED:
			var report = action.report;
			return assign({}, state, {
				[report.ReportId]: report
			});
		case UserConstants.USER_LOGGEDOUT:
			return {};
		default:
			return state;
	}
};
