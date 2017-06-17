var React = require('react');

var ReactBootstrap = require('react-bootstrap');

var Button = ReactBootstrap.Button;
var Panel = ReactBootstrap.Panel;

var StackedBarChart = require('../components/StackedBarChart');
var PieChart = require('../components/PieChart');

var models = require('../models')
var Report = models.Report;
var Tabulation = models.Tabulation;

class ReportsTab extends React.Component {
	constructor() {
		super();
		this.state = {
			initialized: false
		}
		this.onSelectSeries = this.handleSelectSeries.bind(this);
	}
	componentWillMount() {
		this.props.onFetchAllReports();
	}
	componentWillReceiveProps(nextProps) {
		var selected = nextProps.reports.selected;
		if (!this.state.initialized) {
			if (selected == -1 &&
					nextProps.reports.list.length > 0)
				nextProps.onSelectReport(nextProps.reports.map[nextProps.reports.list[0]]);
			this.setState({initialized: true});
		} else if (selected != -1 && !nextProps.reports.tabulations.hasOwnProperty(selected)) {
			nextProps.onTabulateReport(nextProps.reports.map[nextProps.reports.list[0]]);
		} else if (selected != -1 && nextProps.reports.selectedTabulation == null) {
			nextProps.onSelectSeries(nextProps.reports.tabulations[nextProps.reports.list[0]]);
		}
	}
	handleSelectSeries(series) {
		if (series == Tabulation.topLevelSeriesName())
			return;
		var seriesTraversal = this.props.selectedTabulation.seriesTraversal.slice();
		seriesTraversal.push(series);
		var selectedTabulation = this.props.reports.tabulations[this.props.reports.selected];
		this.props.onSelectSeries(selectedTabulation, seriesTraversal);
	}
	render() {
		var selectedTabulation = this.props.reports.selectedTabulation;
		if (!selectedTabulation) {
			return (
				<div></div>
			);
		}

		var titleTracks = [];
		var seriesTraversal = [];

		for (var i = 0; i < this.props.selectedTabulation.seriesTraversal.length; i++) {
			var name = this.props.selectedTabulation.tabulation.Title;
			if (i > 0)
				name = this.props.selectedTabulation.seriesTraversal[i-1];

			// Make a closure for going up the food chain
			var self = this;
			var navOnClick = function() {
				var onSelectTabulation = self.props.onSelectTabulation;
				var report = self.props.reports[self.props.selectedTabulation.tabulation.ReportId];
				var mySeriesTraversal = seriesTraversal.slice();
				return function() {
					onSelectTabulation(report, mySeriesTraversal);
				};
			}();
			titleTracks.push((
				<Button key={i*2} bsStyle="link"
						onClick={navOnClick}>
					{name}
				</Button>
			));
			titleTracks.push((<span key={i*2+1}>/</span>));
			seriesTraversal.push(this.props.selectedTabulation.seriesTraversal[i]);
		}
		if (titleTracks.length == 0) {
			titleTracks.push((
				<Button key={0} bsStyle="link">
					{this.props.selectedTabulation.tabulation.Title}
				</Button>
			));
		} else {
			var i = this.props.selectedTabulation.seriesTraversal.length-1;
			titleTracks.push((
				<Button key={i*2+2} bsStyle="link">
				{this.props.selectedTabulation.seriesTraversal[i]}
				</Button>
			));
		}

		return (
			<Panel header={titleTracks}>
				<PieChart
					report={this.props.selectedTabulation.tabulation}
					onSelectSeries={this.onSelectSeries}
					seriesTraversal={this.props.selectedTabulation.seriesTraversal} />
			</Panel>
		);
	}
}

module.exports = ReportsTab;
