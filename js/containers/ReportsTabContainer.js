var connect = require('react-redux').connect;

var ReportActions = require('../actions/ReportActions');
var ReportsTab = require('../components/ReportsTab');

function mapStateToProps(state) {
	var report_list = [];
	for (var reportId in state.reports.map) {
		if (state.reports.map.hasOwnProperty(reportId))
			report_list.push(state.reports.map[reportId]);
	}
	return {
		reports: state.reports,
		report_list: report_list
	}
}

function mapDispatchToProps(dispatch) {
	return {
		onFetchAllReports: function() {dispatch(ReportActions.fetchAll())},
		onCreateReport: function(report) {dispatch(ReportActions.create(report))},
		onUpdateReport: function(report) {dispatch(ReportActions.update(report))},
		onDeleteReport: function(report) {dispatch(ReportActions.remove(report))},
		onSelectReport: function(report) {dispatch(ReportActions.select(report))},
		onTabulateReport: function(report) {dispatch(ReportActions.tabulate(report))},
		onSelectSeries: function(seriesTraversal) {dispatch(ReportActions.selectSeries(seriesTraversal))}
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(ReportsTab)
