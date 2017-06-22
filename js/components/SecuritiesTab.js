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

class SecurityTemplatePanel extends React.Component {
	constructor() {
		super();
		this.onSearchChange = this.handleSearchChange.bind(this);
	}
	handleSearchChange() {
		this.props.onSearchTemplates(ReactDOM.findDOMNode(this.refs.search).value, 0, this.props.maxResults + 1);
	}
	renderTemplateList() {
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
	}
	render() {
		return (
			<Panel collapsible header="Populate Security from Template...">
				<FormControl type="text"
					placeholder="Search..."
					value={this.props.search}
					onChange={this.onSearchChange}
					ref="search"/>
				{this.renderTemplateList()}
			</Panel>
		);
	}
}

class AddEditSecurityModal extends React.Component {
	getInitialState(props) {
		var s = {
			securityid: -1,
			name: "",
			description: "",
			symbol: "",
			precision: 0,
			type: 1,
			alternateid: ""
		};
		if (props && props.editSecurity != null) {
			s.securityid = props.editSecurity.SecurityId;
			s.name = props.editSecurity.Name;
			s.description = props.editSecurity.Description;
			s.symbol = props.editSecurity.Symbol;
			s.precision = props.editSecurity.Precision;
			s.type = props.editSecurity.Type;
			s.alternateid = props.editSecurity.AlternateId;
		}
		return s;
	}
	constructor() {
		super();
		this.state = this.getInitialState();
		this.onSelectTemplate = this.handleSelectTemplate.bind(this);
		this.onCancel = this.handleCancel.bind(this);
		this.onNameChange = this.handleNameChange.bind(this);
		this.onDescriptionChange = this.handleDescriptionChange.bind(this);
		this.onSymbolChange = this.handleSymbolChange.bind(this);
		this.onPrecisionChange = this.handlePrecisionChange.bind(this);
		this.onTypeChange = this.handleTypeChange.bind(this);
		this.onAlternateIdChange = this.handleAlternateIdChange.bind(this);
		this.onSubmit = this.handleSubmit.bind(this);
	}
	componentWillReceiveProps(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this.getInitialState(nextProps));
		}
	}
	handleSelectTemplate(template) {
		this.setState({
			name: template.Name,
			description: template.Description,
			symbol: template.Symbol,
			precision: template.Precision,
			type: template.Type,
			alternateid: template.AlternateId
		});
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
	handleDescriptionChange() {
		this.setState({
			description: ReactDOM.findDOMNode(this.refs.description).value,
		});
	}
	handleSymbolChange() {
		this.setState({
			symbol: ReactDOM.findDOMNode(this.refs.symbol).value,
		});
	}
	handlePrecisionChange() {
		this.setState({
			precision: +ReactDOM.findDOMNode(this.refs.precision).value,
		});
	}
	handleTypeChange(type) {
		if (type.hasOwnProperty('TypeId'))
			this.setState({
				type: type.TypeId
			});
	}
	handleAlternateIdChange() {
		this.setState({
			alternateid: ReactDOM.findDOMNode(this.refs.alternateid).value,
		});
	}
	handleSubmit() {
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
	}
	render() {
		var headerText = (this.props.editSecurity != null) ? "Edit" : "Create New";
		var buttonText = (this.props.editSecurity != null) ? "Save Changes" : "Create Security";
		var alternateidname = (this.state.type == SecurityType.Currency) ? "ISO 4217 Code" : "CUSIP";
		return (
			<Modal show={this.props.show} onHide={this.onCancel}>
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
						<Col componentClass={ControlLabel} xs={3}>Description</Col>
						<Col xs={9}>
						<FormControl type="text"
							value={this.state.description}
							onChange={this.onDescriptionChange}
							ref="description"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>Symbol or Ticker</Col>
						<Col xs={9}>
						<FormControl type="text"
							value={this.state.symbol}
							onChange={this.onSymbolChange}
							ref="symbol"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>Smallest Fraction Traded</Col>
						<Col xs={9}>
						<FormControl componentClass="select"
							placeholder={this.state.precision}
							value={this.state.precision}
							onChange={this.onPrecisionChange}
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
							onChange={this.onTypeChange}
							ref="type" />
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={3}>{alternateidname}</Col>
						<Col xs={9}>
						<FormControl type="text"
							value={this.state.alternateid}
							onChange={this.onAlternateIdChange}
							ref="alternateid"/>
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

class DeletionFailedModal extends React.Component {
	render() {
		var msg = "We are unable to delete your " + this.props.deletingSecurity.Name + " security because it is in use by " + this.props.securityAccounts.length + " account(s). Please change those accounts to use other securities and try again.";
		if (this.props.user.DefaultCurrency == this.props.deletingSecurity.SecurityId) {
			msg = "We are unable to delete your default currency: " + this.props.deletingSecurity.Name + ". To delete this security, select another as your default currency under Account Settings.";
		}
		return (
			<Modal show={this.props.show} onHide={this.props.onClose}>
				<Modal.Header closeButton>
					<Modal.Title>Cannot Delete Security</Modal.Title>
				</Modal.Header>
				<Modal.Body>
					{msg}
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup className="pull-right">
						<Button onClick={this.props.onClose} bsStyle="warning">Close</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
}

class SecurityList extends React.Component {
	render() {
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
}

class SecuritiesTab extends React.Component {
	constructor() {
		super();
		this.state = {
			creatingNewSecurity: false,
			editingSecurity: false,
			deletionFailedModal: false
		};
		this.onSelectSecurity = this.handleSelectSecurity.bind(this);
		this.onNewSecurity = this.handleNewSecurity.bind(this);
		this.onEditSecurity = this.handleEditSecurity.bind(this);
		this.onDeleteSecurity = this.handleDeleteSecurity.bind(this);
		this.onCreationCancel = this.handleCreationCancel.bind(this);
		this.onCreationSubmit = this.handleCreationSubmit.bind(this);
		this.onEditingCancel = this.handleEditingCancel.bind(this);
		this.onEditingSubmit = this.handleEditingSubmit.bind(this);
		this.onCloseDeletionFailed = this.handleCloseDeletionFailed.bind(this);
	}
	componentWillReceiveProps(nextProps) {
		if (nextProps.selectedSecurity == -1 && nextProps.security_list.length > 0) {
			nextProps.onSelectSecurity(nextProps.security_list[0].SecurityId);
		}
	}
	handleSelectSecurity(security) {
		this.props.onSelectSecurity(security.SecurityId);
	}
	handleNewSecurity() {
		this.setState({creatingNewSecurity: true});
	}
	handleEditSecurity() {
		this.setState({editingSecurity: true});
	}
	handleDeleteSecurity() {
		// check if user has this as their default currency
		var security = this.props.securities[this.props.selectedSecurity];
		if (this.props.selectedSecurityAccounts.length == 0 && security.SecurityId != this.props.user.DefaultCurrency)
			this.props.onDeleteSecurity(security);
		else
			this.setState({deletionFailedModal: true});
	}
	handleCreationCancel() {
		this.setState({creatingNewSecurity: false});
	}
	handleCreationSubmit(security) {
		this.setState({creatingNewSecurity: false});
		this.props.onCreateSecurity(security);
	}
	handleEditingCancel() {
		this.setState({editingSecurity: false});
	}
	handleEditingSubmit(security) {
		this.setState({editingSecurity: false});

		if (security.SecurityId == this.props.user.DefaultCurrency && security.Type != SecurityType.Currency) {
			this.props.onUserError("Unable to modify the default currency to be a non-currency security");
		} else {
			this.props.onUpdateSecurity(security);
		}
	}
	handleCloseDeletionFailed() {
		this.setState({deletionFailedModal: false});
	}
	render() {
		var noSecuritySelected = this.props.selectedSecurity == -1;

		var selectedSecurity = -1;
		if (this.props.securities.hasOwnProperty(this.props.selectedSecurity))
			selectedSecurity = this.props.securities[this.props.selectedSecurity];

		return (
			<div>
			<AddEditSecurityModal
				show={this.state.creatingNewSecurity}
				onCancel={this.onCreationCancel}
				onSubmit={this.onCreationSubmit}
				onSearchTemplates={this.props.onSearchTemplates}
				securityTemplates={this.props.securityTemplates} />
			<AddEditSecurityModal
				show={this.state.editingSecurity}
				editSecurity={selectedSecurity}
				onCancel={this.onEditingCancel}
				onSubmit={this.onEditingSubmit}
				onSearchTemplates={this.props.onSearchTemplates}
				securityTemplates={this.props.securityTemplates} />
			<DeletionFailedModal
				show={this.state.deletionFailedModal}
				user={this.props.user}
				deletingSecurity={selectedSecurity}
				onClose={this.onCloseDeletionFailed}
				securityAccounts={this.props.selectedSecurityAccounts} />
			<ButtonToolbar>
			<ButtonGroup>
				<Button onClick={this.onNewSecurity} bsStyle="success"><Glyphicon glyph='plus-sign'/> New Security</Button>
			</ButtonGroup>
			<ButtonGroup>
				<Combobox
					data={this.props.security_list}
					valueField='SecurityId'
					textField={item => typeof item === 'string' ? item : item.Name + " - " + item.Description}
					value={selectedSecurity}
					onChange={this.onSelectSecurity}
					suggest
					filter='contains'
					ref="security" />
			</ButtonGroup>
			<ButtonGroup>
				<Button onClick={this.onEditSecurity} bsStyle="primary" disabled={noSecuritySelected}><Glyphicon glyph='cog'/> Edit Security</Button>
				<Button onClick={this.onDeleteSecurity} bsStyle="danger" disabled={noSecuritySelected}><Glyphicon glyph='trash'/> Delete Security</Button>
			</ButtonGroup></ButtonToolbar>
			</div>
		);
	}
}

module.exports = SecuritiesTab;
