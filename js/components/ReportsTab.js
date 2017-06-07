var React = require('react');

var ReactBootstrap = require('react-bootstrap');

var Button = ReactBootstrap.Button;
var Panel = ReactBootstrap.Panel;

var StackedBarChart = require('../components/StackedBarChart');

var models = require('../models')
var Report = models.Report;

class ReportsTab extends React.Component {
	constructor() {
		super();
		this.onSelectSeries = this.handleSelectSeries.bind(this);
	}
	componentWillMount() {
		this.props.onFetchReport("monthly_expenses");
	}
	componentWillReceiveProps(nextProps) {
		if (nextProps.reports['monthly_expenses'] && !nextProps.selectedReport.report) {
			this.props.onSelectReport(nextProps.reports['monthly_expenses'], []);
		}
	}
	handleSelectSeries(series) {
		if (series == Report.topLevelAccountName())
			return;
		var seriesTraversal = this.props.selectedReport.seriesTraversal.slice();
		seriesTraversal.push(series);
		this.props.onSelectReport(this.props.reports[this.props.selectedReport.report.ReportId], seriesTraversal);
	}
	render() {
		var report = [];
		if (this.props.selectedReport.report) {
			var titleTracks = [];
			var seriesTraversal = [];

			for (var i = 0; i < this.props.selectedReport.seriesTraversal.length; i++) {
				var name = this.props.selectedReport.report.Title;
				if (i > 0)
					name = this.props.selectedReport.seriesTraversal[i-1];

				// Make a closure for going up the food chain
				var self = this;
				var navOnClick = function() {
					var onSelectReport = self.props.onSelectReport;
					var report = self.props.reports[self.props.selectedReport.report.ReportId];
					var mySeriesTraversal = seriesTraversal.slice();
					return function() {
						onSelectReport(report, mySeriesTraversal);
					};
				}();
				titleTracks.push((
					<Button key={i*2} bsStyle="link"
							onClick={navOnClick}>
						{name}
					</Button>
				));
				titleTracks.push((<span key={i*2+1}>/</span>));
				seriesTraversal.push(this.props.selectedReport.seriesTraversal[i]);
			}
			if (titleTracks.length == 0) {
				titleTracks.push((
					<Button key={0} bsStyle="link">
						{this.props.selectedReport.report.Title}
					</Button>
				));
			} else {
				var i = this.props.selectedReport.seriesTraversal.length-1;
				titleTracks.push((
					<Button key={i*2+2} bsStyle="link">
					{this.props.selectedReport.seriesTraversal[i]}
					</Button>
				));
			}

			report = (<Panel header={titleTracks}>
				<StackedBarChart
					report={this.props.selectedReport.report}
					onSelectSeries={this.onSelectSeries}
					seriesTraversal={this.props.selectedReport.seriesTraversal} />
				</Panel>
			);
		}
		return (
			<div>
				{report}
			</div>
		);
	}
}

module.exports = ReportsTab;
