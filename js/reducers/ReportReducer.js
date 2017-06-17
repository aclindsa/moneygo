var assign = require('object-assign');

var ReportConstants = require('../constants/ReportConstants');
var UserConstants = require('../constants/UserConstants');

const initialState = {
	map: {},
	tabulations: {},
	list: [],
	selected: -1,
	selectedTabulation: null,
	seriesTraversal: []
};

module.exports = function(state = initialState, action) {
	switch (action.type) {
		case ReportConstants.REPORTS_FETCHED:
			var selected = -1;
			var reports = {};
			var list = [];
			for (var i = 0; i < action.reports.length; i++) {
				var report = action.reports[i];
				reports[report.ReportId] = report;
				list.push(report.ReportId);
				if (state.selected == report.ReportId)
					selected = state.selected;
			}
			return assign({}, state, {
				map: reports,
				list: list,
				tabulations: {},
				selected: selected
			});
		case ReportConstants.REPORT_CREATED:
		case ReportConstants.REPORT_UPDATED:
			var report = action.report;
			var reports = assign({}, state.map, {
				[report.ReportId]: report
			});

			var list = [];
			for (var reportId in reports) {
				if (reports.hasOwnProperty(reportId))
					list.push(report.ReportId);
			}
			return assign({}, state, {
				map: reports,
				list: list
			});
		case ReportConstants.REPORT_REMOVED:
			var selected = state.selected;
			if (action.reportId == selected)
				selected = -1;
			var reports = assign({}, state.map);
			delete reports[action.reportId];
			return assign({}, state, {
				map: reports,
				selected: selected
			});
		case ReportConstants.REPORT_SELECTED:
			return assign({}, state, {
				selected: action.report.ReportId,
				selectedTabulation: null,
				seriesTraversal: []
			});
		case ReportConstants.TABULATION_FETCHED:
			var tabulation = action.tabulation;
			return assign({}, state, {
				[tabulation.ReportId]: tabulation
			});
		case ReportConstants.SERIES_SELECTED:
			return {
				selectedTabulation: action.tabulation,
				seriesTraversal: action.seriesTraversal
			};
		case UserConstants.USER_LOGGEDOUT:
			return initialState;
		default:
			return state;
	}
};
