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
		onFetchReport: function(reportname) {dispatch(ReportActions.fetch(reportname))}
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(ReportsTab)
