var assign = require('object-assign');

var ReportConstants = require('../constants/ReportConstants');
var UserConstants = require('../constants/UserConstants');

var models = require('../models.js');
var Tabulation = models.Tabulation;

const initialState = {
	map: {},
	tabulations: {},
	list: [],
	selected: -1,
	selectedTabulation: null,
	seriesTraversal: []
};

function getFlattenedTabulation(tabulation, seriesTraversal) {
		// Descend the tree to the right series to flatten
		var series = tabulation;
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
			flattenedSeries[Tabulation.topLevelSeriesName()] = series.Values;

		var flattenedTabulation = new Tabulation();

		flattenedTabulation.ReportId = tabulation.ReportId;
		flattenedTabulation.Title = tabulation.Title;
		flattenedTabulation.Subtitle = tabulation.Subtitle;
		flattenedTabulation.Units = tabulation.Units;
		flattenedTabulation.Labels = tabulation.Labels.slice();
		flattenedTabulation.FlattenedSeries = flattenedSeries;

		return flattenedTabulation;
}

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
			var selectedTabulation = state.selectedTabulation;
			var seriesTraversal = state.seriesTraversal;
			if (state.selected == action.report.ReportId) {
				selectedTabulation = initialState.selectedTabulation;
				seriesTraversal = initialState.seriesTraversal;
			}

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
				list: list,
				selectedTabulation: selectedTabulation,
				seriesTraversal: seriesTraversal
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
			var selectedTabulation = null;
			if (state.tabulations.hasOwnProperty(action.report.ReportId)) {
				selectedTabulation = getFlattenedTabulation(state.tabulations[action.report.ReportId], initialState.seriesTraversal)
			}
			return assign({}, state, {
				selected: action.report.ReportId,
				selectedTabulation: selectedTabulation,
				seriesTraversal: initialState.seriesTraversal
			});
		case ReportConstants.REPORT_TABULATED:
			var tabulation = action.tabulation;
			var tabulations = assign({}, state.tabulations, {
				[tabulation.ReportId]: tabulation
			});
			var selectedTabulation = state.selectedTabulation;
			var seriesTraversal = state.seriesTraversal;
			if (tabulation.ReportId == state.selected) {
				selectedTabulation = getFlattenedTabulation(tabulation, initialState.seriesTraversal)
				seriesTraversal = initialState.seriesTraversal;
			}
			return assign({}, state, {
				tabulations: tabulations,
				selectedTabulation: selectedTabulation,
				seriesTraversal: seriesTraversal
			});
		case ReportConstants.SERIES_SELECTED:
			return assign({}, state, {
				selectedTabulation: getFlattenedTabulation(state.tabulations[state.selected], action.seriesTraversal),
				seriesTraversal: action.seriesTraversal
			});
		case UserConstants.USER_LOGGEDOUT:
			return initialState;
		default:
			return state;
	}
};
