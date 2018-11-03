var React = require('react');
var ReactDOM = require('react-dom');

var ReactBootstrap = require('react-bootstrap');
var Col = ReactBootstrap.Col;
var Tabs = ReactBootstrap.Tabs;
var Tab = ReactBootstrap.Tab;
var Panel = ReactBootstrap.Panel;
var Modal = ReactBootstrap.Modal;
var Form = ReactBootstrap.Form;
var FormGroup = ReactBootstrap.FormGroup;
var FormControl = ReactBootstrap.FormControl;
var ControlLabel = ReactBootstrap.ControlLabel;
var Checkbox = ReactBootstrap.Checkbox;
var Button = ReactBootstrap.Button;
var ButtonGroup = ReactBootstrap.ButtonGroup;

var Combobox = require('react-widgets').Combobox;

var AccountCombobox = require('./AccountCombobox');

var models = require('../models');
var Account = models.Account;
var AccountType = models.AccountType;
var AccountTypeList = models.AccountTypeList;

class AddEditAccountModal extends React.Component {
	getInitialState(props) {
		var s = {
			accountid: -1,
			security: 1,
			parentaccountid: -1,
			type: 1,
			name: "",
			ofxurl: "",
			ofxorg: "",
			ofxfid: "",
			ofxuser: "",
			ofxbankid: "",
			ofxacctid: "",
			ofxaccttype: "CHECKING",
			ofxclientuid: "",
			ofxappid: "",
			ofxappver: "",
			ofxversion: "",
			ofxnoindent: false,
		};
		if (!props) {
			return s;
		} else if (props.editAccount != null) {
			s.accountid = props.editAccount.AccountId;
			s.name = props.editAccount.Name;
			s.security = props.editAccount.SecurityId;
			s.parentaccountid = props.editAccount.ParentAccountId;
			s.type = props.editAccount.Type;
			s.ofxurl = props.editAccount.OFXURL;
			s.ofxorg = props.editAccount.OFXORG;
			s.ofxfid = props.editAccount.OFXFID;
			s.ofxuser = props.editAccount.OFXUser;
			s.ofxbankid = props.editAccount.OFXBankID;
			s.ofxacctid = props.editAccount.OFXAcctID;
			s.ofxaccttype = props.editAccount.OFXAcctType;
			s.ofxclientuid = props.editAccount.OFXClientUID;
			s.ofxappid = props.editAccount.OFXAppID;
			s.ofxappver = props.editAccount.OFXAppVer;
			s.ofxversion = props.editAccount.OFXVersion;
			s.ofxnoindent = props.editAccount.OFXNoIndent;
		} else if (props.initialParentAccount != null) {
			s.security = props.initialParentAccount.SecurityId;
			s.parentaccountid = props.initialParentAccount.AccountId;
			s.type = props.initialParentAccount.Type;
		}
		return s;
	}
	constructor() {
		super();
		this.state = this.getInitialState();
		this.onCancel = this.handleCancel.bind(this);
		this.onChange = this.handleChange.bind(this);
		this.onNoIndentClick = this.handleNoIndentClick.bind(this);
		this.onSecurityChange = this.handleSecurityChange.bind(this);
		this.onTypeChange = this.handleTypeChange.bind(this);
		this.onParentChange = this.handleParentChange.bind(this);
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
	handleChange() {
		this.setState({
			name: ReactDOM.findDOMNode(this.refs.name).value,
			ofxurl: ReactDOM.findDOMNode(this.refs.ofxurl).value,
			ofxorg: ReactDOM.findDOMNode(this.refs.ofxorg).value,
			ofxfid: ReactDOM.findDOMNode(this.refs.ofxfid).value,
			ofxuser: ReactDOM.findDOMNode(this.refs.ofxuser).value,
			ofxbankid: ReactDOM.findDOMNode(this.refs.ofxbankid).value,
			ofxacctid: ReactDOM.findDOMNode(this.refs.ofxacctid).value,
			ofxclientuid: ReactDOM.findDOMNode(this.refs.ofxclientuid).value,
			ofxappid: ReactDOM.findDOMNode(this.refs.ofxappid).value,
			ofxappver: ReactDOM.findDOMNode(this.refs.ofxappver).value,
			ofxversion: ReactDOM.findDOMNode(this.refs.ofxversion).value,
		});
		if (this.state.type != AccountType.Investment) {
			this.setState({
				ofxaccttype: ReactDOM.findDOMNode(this.refs.ofxaccttype).value,
			});
		}
	}

	handleNoIndentClick() {
		this.setState({ofxnoindent: !this.state.ofxnoindent});
	}
	handleSecurityChange(security) {
		if (security.hasOwnProperty('SecurityId'))
			this.setState({
				security: security.SecurityId
			});
	}
	handleTypeChange(type) {
		if (type.hasOwnProperty('TypeId'))
			this.setState({
				type: type.TypeId
			});
	}
	handleParentChange(parentAccount) {
		this.setState({parentaccountid: parentAccount.AccountId});
	}
	handleSubmit() {
		var a = new Account();

		if (this.props.editAccount != null)
			a.AccountId = this.state.accountid;
		a.Name = this.state.name;
		a.ParentAccountId = this.state.parentaccountid;
		a.SecurityId = this.state.security;
		a.Type = this.state.type;

		a.OFXURL = this.state.ofxurl;
		a.OFXORG = this.state.ofxorg;
		a.OFXFID = this.state.ofxfid;
		a.OFXUser = this.state.ofxuser;
		a.OFXBankID = this.state.ofxbankid;
		a.OFXAcctID = this.state.ofxacctid;
		a.OFXAcctType = this.state.ofxaccttype;
		a.OFXClientUID = this.state.ofxclientuid;
		a.OFXAppID = this.state.ofxappid;
		a.OFXAppVer = this.state.ofxappver;
		a.OFXVersion = this.state.ofxversion;
		a.OFXNoIndent = this.state.ofxnoindent;

		if (this.props.onSubmit != null)
			this.props.onSubmit(a);
	}
	render() {
		var headerText = (this.props.editAccount != null) ? "Edit" : "Create New";
		var buttonText = (this.props.editAccount != null) ? "Save Changes" : "Create Account";
		var rootName = (this.props.editAccount != null) ? "Top-level Account" : "New Top-level Account";
		var ofxBankIdName = "Broker ID";
		var ofxAcctType = [];
		if (this.state.type != AccountType.Investment) {
			ofxBankIdName = "Bank ID";
			ofxAcctType = (
				<FormGroup>
					<Col componentClass={ControlLabel} xs={2}>Account Type</Col>
					<Col xs={10}>
					<FormControl
							componentClass="select"
							placeholder="select"
							value={this.state.ofxaccttype}
							onChange={this.onChange}
							ref="ofxaccttype">
						<option value="CHECKING">Checking</option>
						<option value="SAVINGS">Savings</option>
						<option value="CC">Credit Card</option>
						<option value="MONEYMRKT">Money Market</option>
						<option value="CREDITLINE">Credit Line</option>
						<option value="CD">CD</option>
					</FormControl>
					</Col>
				</FormGroup>
			);
		}
		var bankIdDisabled = (this.state.type != AccountType.Investment && this.state.ofxaccttype == "CC") ? true : false;
		return (
			<Modal show={this.props.show} onHide={this.onCancel}>
				<Modal.Header closeButton>
					<Modal.Title>{headerText} Account</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<Tabs defaultActiveKey={1} id="editAccountTabs">
					<Tab eventKey={1} title="General">
					<Form horizontal onSubmit={this.onSubmit}>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Name</Col>
						<Col xs={10}>
						<FormControl type="text"
							value={this.state.name}
							onChange={this.onChange}
							ref="name"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Parent Account</Col>
						<Col xs={10}>
						<AccountCombobox
							accounts={this.props.accounts}
							accountChildren={this.props.accountChildren}
							value={this.state.parentaccountid}
							rootName={rootName}
							onChange={this.onParentChange}
							ref="parent" />
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Security</Col>
						<Col xs={10}>
						<Combobox
							suggest
							data={this.props.security_list}
							valueField='SecurityId'
							textField={item => typeof item === 'string' ? item : item.Name + " - " + item.Description}
							defaultValue={this.state.security}
							onChange={this.onSecurityChange}
							ref="security" />
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Account Type</Col>
						<Col xs={10}>
						<Combobox
							suggest
							data={AccountTypeList}
							valueField='TypeId'
							textField='Name'
							defaultValue={this.state.type}
							onChange={this.onTypeChange}
							ref="type" />
						</Col>
					</FormGroup>
					</Form>
					</Tab>
					<Tab eventKey={2} title="Sync (OFX)">
					<Form horizontal onSubmit={this.onSubmit}>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>OFX URL</Col>
						<Col xs={10}>
						<FormControl type="text"
							value={this.state.ofxurl}
							onChange={this.onChange}
							ref="ofxurl"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>ORG</Col>
						<Col xs={10}>
						<FormControl type="text"
							value={this.state.ofxorg}
							onChange={this.onChange}
							ref="ofxorg"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>FID</Col>
						<Col xs={10}>
						<FormControl type="text"
							value={this.state.ofxfid}
							onChange={this.onChange}
							ref="ofxfid"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Username</Col>
						<Col xs={10}>
						<FormControl type="text"
							value={this.state.ofxuser}
							onChange={this.onChange}
							ref="ofxuser"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>{ofxBankIdName}</Col>
						<Col xs={10}>
						<FormControl type="text"
							disabled={bankIdDisabled}
							value={this.state.ofxbankid}
							onChange={this.onChange}
							ref="ofxbankid"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Account ID</Col>
						<Col xs={10}>
						<FormControl type="text"
							value={this.state.ofxacctid}
							onChange={this.onChange}
							ref="ofxacctid"/>
						</Col>
					</FormGroup>
					{ofxAcctType}
					<Panel collapsible header="Advanced Settings">
						<FormGroup>
							<Col componentClass={ControlLabel} xs={2}>Client UID</Col>
							<Col xs={10}>
							<FormControl type="text"
								value={this.state.ofxclientuid}
								onChange={this.onChange}
								ref="ofxclientuid"/>
							</Col>
						</FormGroup>
						<FormGroup>
							<Col componentClass={ControlLabel} xs={2}>App ID</Col>
							<Col xs={10}>
							<FormControl type="text"
								value={this.state.ofxappid}
								onChange={this.onChange}
								ref="ofxappid"/>
							</Col>
						</FormGroup>
						<FormGroup>
							<Col componentClass={ControlLabel} xs={2}>App Version</Col>
							<Col xs={10}>
							<FormControl type="text"
								value={this.state.ofxappver}
								onChange={this.onChange}
								ref="ofxappver"/>
							</Col>
						</FormGroup>
						<FormGroup>
							<Col componentClass={ControlLabel} xs={2}>OFX Version</Col>
							<Col xs={10}>
							<FormControl type="text"
								value={this.state.ofxversion}
								onChange={this.onChange}
								ref="ofxversion"/>
							</Col>
						</FormGroup>
						<FormGroup>
							<Col xsOffset={2} xs={10}>
							<Checkbox
								checked={this.state.ofxnoindent ? "checked" : ""}
								onClick={this.onNoIndentClick}>
									Don't indent OFX request files
							</Checkbox>
							</Col>
						</FormGroup>
					</Panel>
					</Form>
					</Tab>
				</Tabs>
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

module.exports = AddEditAccountModal;
