var React = require('react');
var ReactDOM = require('react-dom');

var ReactBootstrap = require('react-bootstrap');
var Grid = ReactBootstrap.Grid;
var Row = ReactBootstrap.Row;
var Col = ReactBootstrap.Col;
var Form = ReactBootstrap.Form;
var FormGroup = ReactBootstrap.FormGroup;
var FormControl = ReactBootstrap.FormControl;
var Checkbox = ReactBootstrap.Checkbox;
var ControlLabel = ReactBootstrap.ControlLabel;
var Button = ReactBootstrap.Button;
var ButtonGroup = ReactBootstrap.ButtonGroup;
var Glyphicon = ReactBootstrap.Glyphicon;
var ListGroup = ReactBootstrap.ListGroup;
var ListGroupItem = ReactBootstrap.ListGroupItem;
var Alert = ReactBootstrap.Alert;
var Modal = ReactBootstrap.Modal;
var Collapse = ReactBootstrap.Collapse;
var Tabs = ReactBootstrap.Tabs;
var Tab = ReactBootstrap.Tab;
var Panel = ReactBootstrap.Panel;

var Combobox = require('react-widgets').Combobox;

var models = require('../models');
var Security = models.Security;
var Account = models.Account;
var AccountType = models.AccountType;
var AccountTypeList = models.AccountTypeList;

var AccountCombobox = require('./AccountCombobox');
var AccountRegister = require('./AccountRegister');

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


class DeleteAccountModal extends React.Component {
	getInitialState(props) {
		if (!props)
			var accountid = -1;
		else if (props.initialAccount != null)
			var accountid = props.initialAccount.AccountId;
		else if (props.accounts.length > 0)
			var accountid = props.accounts[0].AccountId;
		else
			var accountid = -1;

		return {
			error: "",
			accountid: accountid,
			checked: false,
			show: false
		};
	}
	constructor() {
		super();
		this.state = this.getInitialState();
		this.onCancel = this.handleCancel.bind(this);
		this.onChange = this.handleChange.bind(this);
		this.onCheckboxClick = this.handleCheckboxClick.bind(this);
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
	handleChange(account) {
		this.setState({accountid: account.AccountId});
	}
	handleCheckboxClick() {
		this.setState({checked: !this.state.checked});
	}
	handleSubmit() {
		if (this.props.accounts.hasOwnProperty(this.state.accountid)) {
			if (this.state.checked) {
				if (this.props.onSubmit != null)
					this.props.onSubmit(this.props.accounts[this.state.accountid]);
			} else {
				this.setState({error: "You must confirm you wish to delete this account."});
			}
		} else {
			this.setState({error: "You must select an account."});
		}
	}
	render() {
		var checkbox = [];
		if (this.props.accounts.hasOwnProperty(this.state.accountid)) {
			var parentAccountId = this.props.accounts[this.state.accountid].ParentAccountId;
			var parentAccount = "will be deleted and any child accounts will become top-level accounts.";
			if (parentAccountId != -1)
				parentAccount = "and any child accounts will be re-parented to: " + this.props.accounts[parentAccountId].Name;

			var warningString = "I understand that deleting this account cannot be undone and that all transactions " + parentAccount;
			checkbox = (
				<FormGroup>
				<Col xsOffset={2} sm={10}>
				<Checkbox
					checked={this.state.checked ? "checked" : ""}
					onClick={this.onCheckboxClick}>
					{warningString}
				</Checkbox>
				</Col>
				</FormGroup>);
		}
		var warning = [];
		if (this.state.error.length != "") {
			warning = (
				<Alert bsStyle="danger"><strong>Error: </strong>{this.state.error}</Alert>
			);
		}

		return (
			<Modal
				show={this.props.show}
				onHide={this.onCancel}
				ref="modal">
				<Modal.Header closeButton>
					<Modal.Title>Delete Account</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				{warning}
				<Form horizontal onSubmit={this.onSubmit}>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Delete Account</Col>
						<Col xs={10}>
						<AccountCombobox
							includeRoot={false}
							accounts={this.props.accounts}
							accountChildren={this.props.accountChildren}
							value={this.state.accountid}
							onChange={this.onChange}/>
						</Col>
					</FormGroup>
					{checkbox}
				</Form>
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup className="pull-right">
						<Button onClick={this.onCancel} bsStyle="warning">Cancel</Button>
						<Button onClick={this.onSubmit} bsStyle="success">Delete Account</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
}

class AccountTreeNode extends React.Component {
	constructor() {
		super();
		this.state = {expanded: false};
		this.onToggle = this.handleToggle.bind(this);
		this.onChildSelect = this.handleChildSelect.bind(this);
		this.onSelect = this.handleSelect.bind(this);
	}
	handleToggle(e) {
		e.preventDefault();
		this.setState({expanded:!this.state.expanded});
	}
	handleChildSelect(account) {
		if (this.props.onSelect != null)
			this.props.onSelect(account);
	}
	handleSelect() {
		if (this.props.onSelect != null)
			this.props.onSelect(this.props.account);
	}
	render() {
		var glyph = this.state.expanded ? 'minus' : 'plus';
		var active = (this.props.selectedAccount != -1 &&
			this.props.account.AccountId == this.props.selectedAccount);
		var buttonStyle = active ? "info" : "link";

		var self = this;
		var children = this.props.accountChildren[this.props.account.AccountId].map(function(childId) {
			var account = self.props.accounts[childId];
			return (
				<AccountTreeNode
					key={account.AccountId}
					account={account}
					selectedAccount={self.props.selectedAccount}
					accounts={self.props.accounts}
					accountChildren={self.props.accountChildren}
					onSelect={self.onChildSelect}/>
		   );
		});
		var accounttreeClasses = "accounttree"
		var expandButton = [];
		if (children.length > 0) {
			expandButton.push((
				<Button onClick={this.onToggle}
						key={1}
						bsSize="xsmall"
						bsStyle="link"
						className="accounttree-expandbutton">
					<Glyphicon glyph={glyph} bsSize="xsmall"/>
				</Button>
			));
		} else {
			accounttreeClasses += "-nochildren";
		}
		return (
			<div className={accounttreeClasses}>
				{expandButton}
				<Button onClick={this.onSelect}
						bsStyle={buttonStyle}
						className="accounttree-name">
					{this.props.account.Name}
				</Button>
				<Collapse in={this.state.expanded}>
					<div>
						{children}
					</div>
				</Collapse>
			</div>
		);
	}
}

class AccountTree extends React.Component {
	constructor() {
		super();
		this.state = {height: 0};
		this.onSelect = this.handleSelect.bind(this);
	}
	handleSelect(account) {
		if (this.props.onSelect != null) {
			this.props.onSelect(account);
		}
	}
	resize() {
		var div = ReactDOM.findDOMNode(this);
		this.setState({height: div.parentElement.clientHeight - 73});
	}
	componentDidMount() {
		this.resize();
		var self = this;
		$(window).resize(function() {self.resize();});
	}
	render() {
		var accounts = this.props.accounts;

		var children = [];
		for (var accountId in accounts) {
			if (accounts.hasOwnProperty(accountId) &&
					accounts[accountId].isRootAccount()) {
				children.push((<AccountTreeNode
					key={accounts[accountId].AccountId}
					account={accounts[accountId]}
					selectedAccount={this.props.selectedAccount}
					accounts={this.props.accounts}
					accountChildren={this.props.accountChildren}
					onSelect={this.onSelect}/>));
			}
		}

		var style = {height: this.state.height + "px"};

		return (
			<div className="accounttree-root" style={style} >
				{children}
			</div>
		);
	}
}

class AccountsTab extends React.Component {
	constructor() {
		super();
		this.state = {
			creatingNewAccount: false,
			editingAccount: false,
			deletingAccount: false
		};
		this.onNewAccount = this.handleNewAccount.bind(this);
		this.onEditAccount = this.handleEditAccount.bind(this);
		this.onDeleteAccount = this.handleDeleteAccount.bind(this);
		this.onCreationCancel = this.handleCreationCancel.bind(this);
		this.onEditingCancel = this.handleEditingCancel.bind(this);
		this.onDeletionCancel = this.handleDeletionCancel.bind(this);
		this.onCreateAccount = this.handleCreateAccount.bind(this);
		this.onUpdateAccount = this.handleUpdateAccount.bind(this);
		this.onRemoveAccount = this.handleRemoveAccount.bind(this);
		this.onAccountSelected = this.handleAccountSelected.bind(this);
	}
	handleNewAccount() {
		this.setState({creatingNewAccount: true});
	}
	handleEditAccount() {
		this.setState({editingAccount: true});
	}
	handleDeleteAccount() {
		this.setState({deletingAccount: true});
	}
	handleCreationCancel() {
		this.setState({creatingNewAccount: false});
	}
	handleEditingCancel() {
		this.setState({editingAccount: false});
	}
	handleDeletionCancel() {
		this.setState({deletingAccount: false});
	}
	handleCreateAccount(account) {
		if (this.props.onCreateAccount != null)
			this.props.onCreateAccount(account);
		this.setState({creatingNewAccount: false});
	}
	handleUpdateAccount(account) {
		if (this.props.onUpdateAccount != null)
			this.props.onUpdateAccount(account);
		this.setState({editingAccount: false});
	}
	handleRemoveAccount(account) {
		if (this.props.onDeleteAccount != null)
			this.props.onDeleteAccount(account);
		this.setState({deletingAccount: false});
	}
	handleAccountSelected(account) {
		this.props.onSelectAccount(account.AccountId);
		this.props.onFetchTransactionPage(account, 20, 0);
	}
	render() {
		var disabled = (this.props.selectedAccount == -1) ? true : false;

		var selectedAccount = null;
		if (this.props.accounts.hasOwnProperty(this.props.selectedAccount))
			selectedAccount = this.props.accounts[this.props.selectedAccount];

		return (
			<Grid fluid className="fullheight"><Row className="fullheight">
				<Col xs={2} className="fullheight account-column">
					<AddEditAccountModal
						show={this.state.creatingNewAccount}
						initialParentAccount={selectedAccount}
						accounts={this.props.accounts}
						accountChildren={this.props.accountChildren}
						onCancel={this.onCreationCancel}
						onSubmit={this.onCreateAccount}
						security_list={this.props.security_list}/>
					<AddEditAccountModal
						show={this.state.editingAccount}
						editAccount={selectedAccount}
						accounts={this.props.accounts}
						accountChildren={this.props.accountChildren}
						onCancel={this.onEditingCancel}
						onSubmit={this.onUpdateAccount}
						security_list={this.props.security_list}/>
					<DeleteAccountModal
						show={this.state.deletingAccount}
						initialAccount={selectedAccount}
						accounts={this.props.accounts}
						accountChildren={this.props.accountChildren}
						onCancel={this.onDeletionCancel}
						onSubmit={this.onRemoveAccount}/>
					<AccountTree
						accounts={this.props.accounts}
						accountChildren={this.props.accountChildren}
						selectedAccount={this.props.selectedAccount}
						onSelect={this.onAccountSelected}/>
					<ButtonGroup className="account-buttongroup">
						<Button onClick={this.onNewAccount} bsStyle="success">
							<Glyphicon glyph='plus-sign' /></Button>
						<Button onClick={this.onEditAccount}
								bsStyle="primary" disabled={disabled}>
							<Glyphicon glyph='cog' /></Button>
						<Button onClick={this.onDeleteAccount}
								bsStyle="danger" disabled={disabled}>
							<Glyphicon glyph='trash' /></Button>
					</ButtonGroup>
				</Col><Col xs={10} className="fullheight transactions-column">
					<AccountRegister
						pageSize={20}
						selectedAccount={this.props.selectedAccount}
						accounts={this.props.accounts}
						accountChildren={this.props.accountChildren}
						securities={this.props.securities}
						transactions={this.props.transactions}
						transactionPage={this.props.transactionPage}
						imports={this.props.imports}
						onFetchAllAccounts={this.props.onFetchAllAccounts}
						onFetchAllSecurities={this.props.onFetchAllSecurities}
						onCreateTransaction={this.props.onCreateTransaction}
						onUpdateTransaction={this.props.onUpdateTransaction}
						onDeleteTransaction={this.props.onDeleteTransaction}
						onSelectTransaction={this.props.onSelectTransaction}
						onUnselectTransaction={this.props.onUnselectTransaction}
						onFetchTransactionPage={this.props.onFetchTransactionPage}
						onOpenImportModal={this.props.onOpenImportModal}
						onCloseImportModal={this.props.onCloseImportModal}
						onImportOFX={this.props.onImportOFX}
						onImportOFXFile={this.props.onImportOFXFile}
						onImportGnucash={this.props.onImportGnucash} />
				</Col>
			</Row></Grid>
		);
	}
}

module.exports = AccountsTab;
