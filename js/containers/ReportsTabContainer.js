var connect = require('react-redux').connect;

var ReportActions = require('../actions/ReportActions');
var ReportsTab = require('../components/ReportsTab');

function mapStateToProps(state) {
	return {
		reports: state.reports
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
		onSelectSeries: function(tabulation, seriesTraversal) {dispatch(ReportActions.selectSeries(tabulation, seriesTraversal))}
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(ReportsTab)
