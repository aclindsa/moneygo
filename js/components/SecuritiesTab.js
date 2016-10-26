var React = require('react');
var ReactDOM = require('react-dom');

var ReactBootstrap = require('react-bootstrap');
var Grid = ReactBootstrap.Grid;
var Row = ReactBootstrap.Row;
var Col = ReactBootstrap.Col;
var Form = ReactBootstrap.Form;
var FormGroup = ReactBootstrap.FormGroup;
var FormControl = ReactBootstrap.FormControl;
var ControlLabel = ReactBootstrap.ControlLabel;
var Button = ReactBootstrap.Button;
var ButtonGroup = ReactBootstrap.ButtonGroup;
var ButtonToolbar = ReactBootstrap.ButtonToolbar;
var Glyphicon = ReactBootstrap.Glyphicon;
var ListGroup = ReactBootstrap.ListGroup;
var ListGroupItem = ReactBootstrap.ListGroupItem;
var Modal = ReactBootstrap.Modal;
var Panel = ReactBootstrap.Panel;

var Combobox = require('react-widgets').Combobox;

var models = require('../models');
var Security = models.Security;
var SecurityType = models.SecurityType;
var SecurityTypeList = models.SecurityTypeList;

const SecurityTemplatePanel = React.createClass({
	handleSearchChange: function(){
		this.props.onSearchTemplates(ReactDOM.findDOMNode(this.refs.search).value, 0, this.props.maxResults + 1);
	},
	renderTemplateList: function() {
		var templates = this.props.securityTemplates;
		if (this.props.search != "") {
			var items = [];
			for (var i = 0; i < templates.length && i < 15; i++) {
				var template = templates[i];
				var self = this;
				var onClickFn = (function() {
					var j = i;
					return function(){self.props.onSelectTemplate(templates[j])};
				})();
				var key = template.Type.toString() + template.AlternateId;
				items.push((
					<ListGroupItem onClick={onClickFn} key={key}>
						{template.Name} - {template.Description}
					</ListGroupItem>
				));
			}
			if (templates.length > this.props.maxResults) {
				items.push((
					<ListGroupItem disabled key="too-many-templates">
						Too many templates to display, please refine your search...
					</ListGroupItem>
				));
			} else if (templates.length == 0) {
				items.push((
					<ListGroupItem disabled key="no-templates">
						Sorry, no templates matched your search...
					</ListGroupItem>
				));
			}
			return (
				<div>
					<br />
					<ControlLabel>Select a template to populate your security:</ControlLabel>
					<ListGroup>
					{items}
					</ListGroup>
				</div>
			);
		}
	},
	render: function() {
		return (
			<Panel collapsible header="Populate Security from Template...">
				<FormControl type="text"
					placeholder="Search..."
					value={this.props.search}
					onChange={this.handleSearchChange}
					ref="search"/>
				{this.renderTemplateList()}
			</Panel>
		);
	}
});

const AddEditSecurityModal = React.createClass({
	getInitialState: function() {
		var s = {
			securityid: -1,
			name: "",
			description: "",
			symbol: "",
			precision: 0,
			type: 1,
			alternateid: ""
		};
		if (this.props.editSecurity != null) {
			s.securityid = this.props.editSecurity.SecurityId;
			s.name = this.props.editSecurity.Name;
			s.description = this.props.editSecurity.Description;
			s.symbol = this.props.editSecurity.Symbol;
			s.precision = this.props.editSecurity.Precision;
			s.type = this.props.editSecurity.Type;
			s.alternateid = this.props.editSecurity.AlternateId;
		}
		return s;
	},
	onSelectTemplate: function(template) {
		this.setState({
			name: template.Name,
			description: template.Description,
			symbol: template.Symbol,
			precision: template.Precision,
			type: template.Type,
			alternateid: template.AlternateId 
		});
	},
	handleCancel: function() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	},
	handleNameChange: function() {
		this.setState({
			name: ReactDOM.findDOMNode(this.refs.name).value,
		});
	},
	handleDescriptionChange: function() {
		this.setState({
			description: ReactDOM.findDOMNode(this.refs.description).value,
		});
	},
	handleSymbolChange: function() {
		this.setState({
			symbol: ReactDOM.findDOMNode(this.refs.symbol).value,
		});
	},
	handlePrecisionChange: function() {
		this.setState({
			precision: +ReactDOM.findDOMNode(this.refs.precision).value,
		});
	},
	handleTypeChange: function(type) {
		if (type.hasOwnProperty('TypeId'))
			this.setState({
				type: type.TypeId
			});
	},
	handleAlternateIdChange: function() {
		this.setState({
			alternateid: ReactDOM.findDOMNode(this.refs.alternateid).value,
		});
	},
	handleSubmit: function() {
		var s = new Security();

		if (this.props.editSecurity != null)
			s.SecurityId = this.state.securityid;
		s.Name = this.state.name;
		s.Description = this.state.description;
		s.Symbol = this.state.symbol;
		s.Precision = this.state.precision;
		s.Type = this.state.type;
		s.AlternateId = this.state.alternateid;

		if (this.props.onSubmit != null)
			this.props.onSubmit(s);
	},
	componentWillReceiveProps: function(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this.getInitialState());
		}
	},
	render: function() {
		var headerText = (this.props.editSecurity != null) ? "Edit" : "Create New";
		var buttonText = (this.props.editSecurity != null) ? "Save Changes" : "Create Security";
		var alternateidname = (this.state.type == SecurityType.Currency) ? "ISO 4217 Code" : "CUSIP";
		return (
			<Modal show={this.props.show} onHide={this.handleCancel}>
				<Modal.Header closeButton>
					<Modal.Title>{headerText} Security</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<SecurityTemplatePanel
					search={this.props.securityTemplates.search}
					securityTemplates={this.props.securityTemplates.templates}
					onSearchTemplates={this.props.onSearchTemplates}
					maxResults={15}
					onSelectTemplate={this.onSelectTemplate} />
				<Form horizontal onSubmit={this.handleSubmit}>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>Name</Col>
						<Col xs={9}>
						<FormControl type="text"
							value={this.state.name}
							onChange={this.handleNameChange}
							ref="name"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>Description</Col>
						<Col xs={9}>
						<FormControl type="text"
							value={this.state.description}
							onChange={this.handleDescriptionChange}
							ref="description"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>Symbol or Ticker</Col>
						<Col xs={9}>
						<FormControl type="text"
							value={this.state.symbol}
							onChange={this.handleSymbolChange}
							ref="symbol"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>Smallest Fraction Traded</Col>
						<Col xs={9}>
						<FormControl componentClass="select"
							placeholder={this.state.precision}
							value={this.state.precision}
							onChange={this.handlePrecisionChange}
							ref="precision">
								<option value={0}>1</option>
								<option value={1}>0.1 (1/10)</option>
								<option value={2}>0.01 (1/100)</option>
								<option value={3}>0.001 (1/1000)</option>
								<option value={4}>0.0001 (1/10000)</option>
								<option value={5}>0.00001 (1/100000)</option>
						</FormControl>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>Security Type</Col>
						<Col xs={9}>
						<Combobox
							suggest
							data={SecurityTypeList}
							valueField='TypeId'
							textField='Name'
							value={this.state.type}
							onChange={this.handleTypeChange}
							ref="type" />
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>{alternateidname}</Col>
						<Col xs={9}>
						<FormControl type="text"
							value={this.state.alternateid}
							onChange={this.handleAlternateIdChange}
							ref="alternateid"/>
						</Col>
					</FormGroup>
				</Form>
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup className="pull-right">
						<Button onClick={this.handleCancel} bsStyle="warning">Cancel</Button>
						<Button onClick={this.handleSubmit} bsStyle="success">{buttonText}</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
});

const SecurityList = React.createClass({
	render: function() {
		var children = [];
		var self = this;
		for (var securityId in this.props.securities) {
			if (this.props.securities.hasOwnProperty(securityId)) {
				var buttonStyle = (securityId == this.props.selectedSecurity) ? "info" : "link";
				var onClickFn = (function() {
					var id = securityId;
					return function(){self.props.onSelectSecurity(id)};
				})();
				children.push((<Button
						bsStyle={buttonStyle}
						key={securityId}
						onClick={onClickFn}>
					{this.props.securities[securityId].Name} - {this.props.securities[securityId].Description}
				</Button>));
			}
		}

		return (
			<div>
				{children}
			</div>
		);
	}
});

module.exports = React.createClass({
	displayName: "SecuritiesTab",
	getInitialState: function() {
		return {
			creatingNewSecurity: false,
			editingSecurity: false
		};
	},
	handleNewSecurity: function() {
		this.setState({creatingNewSecurity: true});
	},
	handleEditSecurity: function() {
		this.setState({editingSecurity: true});
	},
	handleCreationCancel: function() {
		this.setState({creatingNewSecurity: false});
	},
	handleCreationSubmit: function(security) {
		this.setState({creatingNewSecurity: false});
		this.props.onCreateSecurity(security);
	},
	handleEditingCancel: function() {
		this.setState({editingSecurity: false});
	},
	handleEditingSubmit: function(security) {
		this.setState({editingSecurity: false});
		this.props.onUpdateSecurity(security);
	},
	render: function() {
		var editDisabled = this.props.selectedSecurity == -1;

		var selectedSecurity = null;
		if (this.props.securities.hasOwnProperty(this.props.selectedSecurity))
			selectedSecurity = this.props.securities[this.props.selectedSecurity];

		return (
			<Grid fluid className="fullheight"><Row className="fullheight">
				<Col xs={3} className="fullheight securitylist-column">
				<AddEditSecurityModal
					show={this.state.creatingNewSecurity}
					onCancel={this.handleCreationCancel}
					onSubmit={this.handleCreationSubmit}
					onSearchTemplates={this.props.onSearchTemplates}
					securityTemplates={this.props.securityTemplates} />
				<AddEditSecurityModal
					show={this.state.editingSecurity}
					editSecurity={selectedSecurity}
					onCancel={this.handleEditingCancel}
					onSubmit={this.handleEditingSubmit}
					onSearchTemplates={this.props.onSearchTemplates}
					securityTemplates={this.props.securityTemplates} />
				<SecurityList
					selectedSecurity={this.props.selectedSecurity}
					securities={this.props.securities}
					onSelectSecurity={this.props.onSelectSecurity} />
				</Col><Col xs={9} className="fullheight securities-column">
					<ButtonToolbar className="pull-right"><ButtonGroup>
						<Button onClick={this.handleEditSecurity} bsStyle="primary" disabled={editDisabled}><Glyphicon glyph='cog'/> Edit Security</Button>
						<Button onClick={this.handleNewSecurity} bsStyle="success"><Glyphicon glyph='plus-sign'/> New Security</Button>
					</ButtonGroup></ButtonToolbar>
				</Col>
			</Row></Grid>
		);
	}
});
