var ReportConstants = require('../constants/ReportConstants');

var ErrorActions = require('./ErrorActions');

var models = require('../models.js');
var Report = models.Report;
var Tabulation = models.Tabulation;
var Error = models.Error;

function fetchReports() {
	return {
		type: ReportConstants.FETCH_REPORTS
	}
}

function reportsFetched(reports) {
	return {
		type: ReportConstants.REPORTS_FETCHED,
		reports: reports
	}
}

function createReport() {
	return {
		type: ReportConstants.CREATE_REPORT
	}
}

function reportCreated(report) {
	return {
		type: ReportConstants.REPORT_CREATED,
		report: report
	}
}

function updateReport() {
	return {
		type: ReportConstants.UPDATE_REPORT
	}
}

function reportUpdated(report) {
	return {
		type: ReportConstants.REPORT_UPDATED,
		report: report
	}
}

function removeReport() {
	return {
		type: ReportConstants.REMOVE_REPORT
	}
}

function reportRemoved(reportId) {
	return {
		type: ReportConstants.REPORT_REMOVED,
		reportId: reportId
	}
}

function reportSelected(report) {
	return {
		type: ReportConstants.REPORT_SELECTED,
		report: report
	}
}

function tabulateReport(report) {
	return {
		type: ReportConstants.TABULATE_REPORT,
		report: report
	}
}

function reportTabulated(report, tabulation) {
	return {
		type: ReportConstants.REPORT_TABULATED,
		report: report,
		tabulation: tabulation
	}
}

function selectionCleared() {
	return {
		type: ReportConstants.SELECTION_CLEARED
	}
}

function seriesSelected(seriesTraversal) {
	return {
		type: ReportConstants.SERIES_SELECTED,
		seriesTraversal: seriesTraversal
	}
}

function fetchAll() {
	return function (dispatch) {
		dispatch(fetchReports());

		$.ajax({
			type: "GET",
			dataType: "json",
			url: "report/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					dispatch(reportsFetched(data.reports.map(function(json) {
						var r = new Report();
						r.fromJSON(json);
						return r;
					})));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function create(report) {
	return function (dispatch) {
		dispatch(createReport());

		$.ajax({
			type: "POST",
			dataType: "json",
			url: "report/",
			data: {report: report.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					var r = new Report();
					r.fromJSON(data);
					dispatch(reportCreated(r));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function update(report) {
	return function (dispatch) {
		dispatch(updateReport());

		$.ajax({
			type: "PUT",
			dataType: "json",
			url: "report/"+report.ReportId+"/",
			data: {report: report.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					var r = new Report();
					r.fromJSON(data);
					dispatch(reportUpdated(r));
					dispatch(tabulate(r));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function remove(report) {
	return function(dispatch) {
		dispatch(removeReport());

		$.ajax({
			type: "DELETE",
			dataType: "json",
			url: "report/"+report.ReportId+"/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					dispatch(reportRemoved(report.ReportId));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function tabulate(report) {
	return function (dispatch) {
		dispatch(tabulateReport(report));

		$.ajax({
			type: "GET",
			dataType: "json",
			url: "report/"+report.ReportId+"/tabulation/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					var t = new Tabulation();
					t.fromJSON(data);
					dispatch(reportTabulated(report, t));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

module.exports = {
	fetchAll: fetchAll,
	create: create,
	update: update,
	remove: remove,
	tabulate: tabulate,
	select: reportSelected,
	selectSeries: seriesSelected
};
