var ReportConstants = require('../constants/ReportConstants');

var ErrorActions = require('./ErrorActions');

var models = require('../models.js');
var Report = models.Report;
var Error = models.Error;

function fetchReport(reportName) {
	return {
		type: ReportConstants.FETCH_REPORT,
		reportName: reportName
	}
}

function reportFetched(report) {
	return {
		type: ReportConstants.REPORT_FETCHED,
		report: report
	}
}

function selectReport(report, seriesTraversal) {
	return {
		type: ReportConstants.SELECT_REPORT,
		report: report,
		seriesTraversal: seriesTraversal
	}
}

function reportSelected(flattenedReport, seriesTraversal) {
	return {
		type: ReportConstants.REPORT_SELECTED,
		report: flattenedReport,
		seriesTraversal: seriesTraversal
	}
}

function fetch(report) {
	return function (dispatch) {
		dispatch(fetchReport(report));

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

function select(report, seriesTraversal) {
	return function (dispatch) {
		if (!seriesTraversal)
			seriesTraversal = [];
		dispatch(selectReport(report, seriesTraversal));

		// Descend the tree to the right series to flatten
		var series = report;
		for (var i=0; i < seriesTraversal.length; i++) {
			if (!series.Series.hasOwnProperty(seriesTraversal[i])) {
				dispatch(ErrorActions.clientError("Invalid series"));
				return;
			}
			series = series.Series[seriesTraversal[i]];
		}

		// Actually flatten the data
		var flattenedSeries = series.mapReduceChildren(null,
			function(accumulator, currentValue, currentIndex, array) {
				return accumulator + currentValue;
			}
		);

		// Add back in any values from the current level
		if (series.hasOwnProperty('Values'))
			flattenedSeries[report.topLevelAccountName] = series.Values;

		var flattenedReport = new Report();

		flattenedReport.ReportId = report.ReportId;
		flattenedReport.Title = report.Title;
		flattenedReport.Subtitle = report.Subtitle;
		flattenedReport.XAxisLabel = report.XAxisLabel;
		flattenedReport.YAxisLabel = report.YAxisLabel;
		flattenedReport.Labels = report.Labels.slice();
		flattenedReport.FlattenedSeries = flattenedSeries;

		dispatch(reportSelected(flattenedReport, seriesTraversal));
	};
}

module.exports = {
	fetch: fetch,
	select: select
};
