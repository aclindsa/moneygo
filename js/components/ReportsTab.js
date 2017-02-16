var React = require('react');

var StackedBarChart = require('../components/StackedBarChart');

module.exports = React.createClass({
	displayName: "ReportsTab",
	getInitialState: function() {
		return { };
	},
	componentWillMount: function() {
		this.props.onFetchReport("monthly_expenses");
	},
	render: function() {
		report = [];
		if (this.props.reports['monthly_expenses'])
			report = (<StackedBarChart data={this.props.reports['monthly_expenses']} />);
		return (
			<div>
				{report}
			</div>
		);
	}
});
