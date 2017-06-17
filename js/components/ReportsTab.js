var React = require('react');
var ReactDOM = require('react-dom');

var ReactBootstrap = require('react-bootstrap');

var Col = ReactBootstrap.Col;
var Form = ReactBootstrap.Form;
var FormGroup = ReactBootstrap.FormGroup;
var FormControl = ReactBootstrap.FormControl;
var ControlLabel = ReactBootstrap.ControlLabel;
var Button = ReactBootstrap.Button;
var ButtonGroup = ReactBootstrap.ButtonGroup;
var ButtonToolbar = ReactBootstrap.ButtonToolbar;
var Glyphicon = ReactBootstrap.Glyphicon;
var Panel = ReactBootstrap.Panel;
var Modal = ReactBootstrap.Modal;
var ProgressBar = ReactBootstrap.ProgressBar;

var Combobox = require('react-widgets').Combobox;

var CodeMirror = require('react-codemirror');
require('codemirror/mode/lua/lua');

var StackedBarChart = require('../components/StackedBarChart');
var PieChart = require('../components/PieChart');

var models = require('../models')
var Report = models.Report;
var Tabulation = models.Tabulation;

class AddEditReportModal extends React.Component {
	getInitialState(props) {
		var s = {
			reportid: -1,
			name: "",
			lua: ""
		};
		if (props && props.editReport != null) {
			s.reportid = props.editReport.ReportId;
			s.name = props.editReport.Name;
			s.lua = props.editReport.Lua;
		}
		return s;
	}
	constructor() {
		super();
		this.state = this.getInitialState();
		this.onCancel = this.handleCancel.bind(this);
		this.onNameChange = this.handleNameChange.bind(this);
		this.onLuaChange = this.handleLuaChange.bind(this);
		this.onSubmit = this.handleSubmit.bind(this);
	}
	componentWillReceiveProps(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this.getInitialState(nextProps));
		}
	}
	handleCancel() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	}
	handleNameChange() {
		this.setState({
			name: ReactDOM.findDOMNode(this.refs.name).value,
		});
	}
	handleLuaChange(lua) {
		this.setState({
			lua: lua
		});
	}
	handleSubmit() {
		var r = new Report();

		if (this.props.editReport != null)
			r.ReportId = this.state.reportid;
		r.Name = this.state.name;
		r.Lua = this.state.lua;

		if (this.props.onSubmit != null)
			this.props.onSubmit(r);
	}
	render() {
		var headerText = (this.props.editReport != null) ? "Edit" : "Create New";
		var buttonText = (this.props.editReport != null) ? "Save Changes" : "Create Report";

		var codeMirrorOptions = {
			lineNumbers: true,
			mode: 'lua',
		};
		return (
			<Modal show={this.props.show} onHide={this.onCancel} bsSize="large">
				<Modal.Header closeButton>
					<Modal.Title>{headerText} Report</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<Form horizontal onSubmit={this.onSubmit}>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>Name</Col>
						<Col xs={9}>
						<FormControl type="text"
							value={this.state.name}
							onChange={this.onNameChange}
							ref="name"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>Lua Code</Col>
						<Col xs={9}>
						<CodeMirror
							value={this.state.lua}
							onChange={this.onLuaChange}
							options={codeMirrorOptions} />
						</Col>
					</FormGroup>
				</Form>
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup className="pull-right">
						<Button onClick={this.onCancel} bsStyle="warning">Cancel</Button>
						<Button onClick={this.onSubmit} bsStyle="success">{buttonText}</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
}

class ReportsTab extends React.Component {
	constructor() {
		super();
		this.state = {
			initialized: false,
			creatingNewReport: false,
			editingReport: false
		}
		this.onSelectSeries = this.handleSelectSeries.bind(this);
		this.onSelectReport = this.handleSelectReport.bind(this);
		this.onNewReport = this.handleNewReport.bind(this);
		this.onEditReport = this.handleEditReport.bind(this);
		this.onDeleteReport = this.handleDeleteReport.bind(this);
		this.onCreationCancel = this.handleCreationCancel.bind(this);
		this.onCreationSubmit = this.handleCreationSubmit.bind(this);
		this.onEditingCancel = this.handleEditingCancel.bind(this);
		this.onEditingSubmit = this.handleEditingSubmit.bind(this);
	}
	componentWillMount() {
		this.props.onFetchAllReports();
	}
	componentWillReceiveProps(nextProps) {
		var selected = nextProps.reports.selected;
		if (!this.state.initialized) {
			if (selected == -1 &&
					nextProps.reports.list.length > 0) {
				nextProps.onSelectReport(nextProps.reports.map[nextProps.reports.list[0]]);
				nextProps.onTabulateReport(nextProps.reports.map[nextProps.reports.list[0]]);
				this.setState({initialized: true});
			}
		}
	}
	handleSelectSeries(series) {
		if (series == Tabulation.topLevelSeriesName())
			return;
		var seriesTraversal = this.props.reports.seriesTraversal.slice();
		seriesTraversal.push(series);
		this.props.onSelectSeries(seriesTraversal);
	}
	handleSelectReport(report) {
		this.props.onSelectReport(report);
		if (!this.props.reports.tabulations.hasOwnProperty(report.ReportId))
			this.props.onTabulateReport(report);
	}
	handleNewReport() {
		this.setState({creatingNewReport: true});
	}
	handleEditReport() {
		this.setState({editingReport: true});
	}
	handleDeleteReport() {
		this.props.onDeleteReport(this.props.reports.map[this.props.reports.selected]);
	}
	handleCreationCancel() {
		this.setState({creatingNewReport: false});
	}
	handleCreationSubmit(report) {
		this.setState({creatingNewReport: false});
		this.props.onCreateReport(report);
	}
	handleEditingCancel() {
		this.setState({editingReport: false});
	}
	handleEditingSubmit(report) {
		this.setState({editingReport: false});
		this.props.onUpdateReport(report);
	}
	render() {
		var selectedTabulation = this.props.reports.selectedTabulation;
		var reportPanel = [];
		if (selectedTabulation) {
			var titleTracks = [];
			var seriesTraversal = [];

			for (var i = 0; i < this.props.reports.seriesTraversal.length; i++) {
				var name = this.props.reports.selectedTabulation.Title;
				if (i > 0)
					name = this.props.reports.seriesTraversal[i-1];

				// Make a closure for going up the food chain
				var self = this;
				var navOnClick = function() {
					var onSelectSeries = self.props.onSelectSeries;
					var mySeriesTraversal = seriesTraversal.slice();
					return function() {
						onSelectSeries(mySeriesTraversal);
					};
				}();
				titleTracks.push((
					<Button key={i*2} bsStyle="link"
							onClick={navOnClick}>
						{name}
					</Button>
				));
				titleTracks.push((<span key={i*2+1}>/</span>));
				seriesTraversal.push(this.props.reports.seriesTraversal[i]);
			}
			if (titleTracks.length == 0) {
				titleTracks.push((
					<Button key={0} bsStyle="link">
						{this.props.reports.selectedTabulation.Title}
					</Button>
				));
			} else {
				var i = this.props.reports.seriesTraversal.length-1;
				titleTracks.push((
					<Button key={i*2+2} bsStyle="link">
					{this.props.reports.seriesTraversal[i]}
					</Button>
				));
			}

			if (this.props.reports.selectedTabulation.Labels.length > 1)
				var report = (
					<StackedBarChart
						report={this.props.reports.selectedTabulation}
						onSelectSeries={this.onSelectSeries}
						seriesTraversal={this.props.reports.seriesTraversal} />
				);
			else
				var report = (
					<PieChart
						report={this.props.reports.selectedTabulation}
						onSelectSeries={this.onSelectSeries}
						seriesTraversal={this.props.reports.seriesTraversal} />
				);

			reportPanel = (
				<Panel header={titleTracks}>
					{report}
				</Panel>
			);
		} else if (this.props.reports.selected != -1) {
			reportPanel = (
				<Panel header={this.props.reports.map[this.props.reports.selected].Name}>
					<ProgressBar active now={100} label={"Tabulating Report..."} />
				</Panel>
			);
		}

		var noReportSelected = this.props.reports.selected == -1;
		var selectedReport = -1;
		if (this.props.reports.map.hasOwnProperty(this.props.reports.selected))
			selectedReport = this.props.reports.map[this.props.reports.selected];

		return (
			<div>
			<AddEditReportModal
				show={this.state.creatingNewReport}
				onCancel={this.onCreationCancel}
				onSubmit={this.onCreationSubmit} />
			<AddEditReportModal
				show={this.state.editingReport}
				editReport={selectedReport}
				onCancel={this.onEditingCancel}
				onSubmit={this.onEditingSubmit} />
			<ButtonToolbar>
			<ButtonGroup>
				<Button onClick={this.onNewReport} bsStyle="success"><Glyphicon glyph='plus-sign'/> New Report</Button>
			</ButtonGroup>
			<ButtonGroup>
				<Combobox
					data={this.props.report_list}
					valueField='ReportId'
					textField={item => typeof item === 'string' ? item : item.Name}
					value={selectedReport}
					onChange={this.onSelectReport}
					suggest
					filter='contains'
					ref="report" />
			</ButtonGroup>
			<ButtonGroup>
				<Button onClick={this.onEditReport} bsStyle="primary" disabled={noReportSelected}><Glyphicon glyph='cog'/> Edit Report</Button>
				<Button onClick={this.onDeleteReport} bsStyle="danger" disabled={noReportSelected}><Glyphicon glyph='trash'/> Delete Report</Button>
			</ButtonGroup></ButtonToolbar>
			{reportPanel}
			</div>
		);
	}
}

module.exports = ReportsTab;
