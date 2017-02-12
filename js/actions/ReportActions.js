var ReportConstants = require('../constants/ReportConstants');

var ErrorActions = require('./ErrorActions');

var models = require('../models.js');
var Report = models.Report;
var Error = models.Error;

function fetchReport() {
	return {
		type: ReportConstants.FETCH_REPORT
	}
}

function reportFetched(report) {
	return {
		type: ReportConstants.REPORT_FETCHED,
		report: report
	}
}

function fetch(report) {
	return function (dispatch) {
		dispatch(fetchReport());

		$.ajax({
			type: "GET",
			dataType: "json",
			url: "report/"+report+"/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					var r = new Report();
					r.fromJSON(data);
					dispatch(reportFetched(r));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

module.exports = {
	fetch: fetch
};
