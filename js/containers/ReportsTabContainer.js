var connect = require('react-redux').connect;

var ReportActions = require('../actions/ReportActions');
var ReportsTab = require('../components/ReportsTab');

function mapStateToProps(state) {
	return {
		reports: state.reports,
		selectedReport: state.selectedReport
	}
}

function mapDispatchToProps(dispatch) {
	return {
		onFetchReport: function(reportname) {dispatch(ReportActions.fetch(reportname))},
		onSelectReport: function(report, seriesTraversal) {dispatch(ReportActions.select(report, seriesTraversal))}
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(ReportsTab)
