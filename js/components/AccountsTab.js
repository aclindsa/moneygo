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

var Combobox = require('react-widgets').Combobox;

var models = require('../models');
var Security = models.Security;
var Account = models.Account;
var AccountTypeList = models.AccountTypeList;

var AccountCombobox = require('./AccountCombobox');
var AccountRegister = require('./AccountRegister');

const AddEditAccountModal = React.createClass({
	getInitialState: function() {
		var s = {
			accountid: -1,
			security: 1,
			parentaccountid: -1,
			type: 1,
			name: ""
		};
		if (this.props.editAccount != null) {
			s.accountid = this.props.editAccount.AccountId;
			s.name = this.props.editAccount.Name;
			s.security = this.props.editAccount.SecurityId;
			s.parentaccountid = this.props.editAccount.ParentAccountId;
			s.type = this.props.editAccount.Type;
		} else if (this.props.initialParentAccount != null) {
			s.security = this.props.initialParentAccount.SecurityId;
			s.parentaccountid = this.props.initialParentAccount.AccountId;
			s.type = this.props.initialParentAccount.Type;
		}
		return s;
	},
	handleCancel: function() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	},
	handleChange: function() {
		this.setState({
			name: ReactDOM.findDOMNode(this.refs.name).value,
		});
	},
	handleSecurityChange: function(security) {
		if (security.hasOwnProperty('SecurityId'))
			this.setState({
				security: security.SecurityId
			});
	},
	handleTypeChange: function(type) {
		if (type.hasOwnProperty('TypeId'))
			this.setState({
				type: type.TypeId
			});
	},
	handleParentChange: function(parentAccount) {
		this.setState({parentaccountid: parentAccount.AccountId});
	},
	handleSubmit: function() {
		var a = new Account();

		if (this.props.editAccount != null)
			a.AccountId = this.state.accountid;
		a.Name = this.state.name;
		a.ParentAccountId = this.state.parentaccountid;
		a.SecurityId = this.state.security;
		a.Type = this.state.type;

		if (this.props.onSubmit != null)
			this.props.onSubmit(a);
	},
	componentWillReceiveProps: function(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this.getInitialState());
		}
	},
	render: function() {
		var headerText = (this.props.editAccount != null) ? "Edit" : "Create New";
		var buttonText = (this.props.editAccount != null) ? "Save Changes" : "Create Account";
		var rootName = (this.props.editAccount != null) ? "Top-level Account" : "New Top-level Account";
		return (
			<Modal show={this.props.show} onHide={this.handleCancel}>
				<Modal.Header closeButton>
					<Modal.Title>{headerText} Account</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<Form horizontal onSubmit={this.handleSubmit}>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Name</Col>
						<Col xs={10}>
						<FormControl type="text"
							value={this.state.name}
							onChange={this.handleChange}
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
							onChange={this.handleParentChange}
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
							onChange={this.handleSecurityChange}
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
							onChange={this.handleTypeChange}
							ref="type" />
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


const DeleteAccountModal = React.createClass({
	getInitialState: function() {
		if (this.props.initialAccount != null)
			var accountid = this.props.initialAccount.AccountId;
		else if (this.props.accounts.length > 0)
			var accountid = this.props.accounts[0].AccountId;
		else
			var accountid = -1;
		return {error: "",
			accountid: accountid,
			checked: false,
			show: false};
	},
	handleCancel: function() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	},
	handleChange: function(account) {
		this.setState({accountid: account.AccountId});
	},
	handleCheckboxClick: function() {
		this.setState({checked: !this.state.checked});
	},
	handleSubmit: function() {
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
	},
	componentWillReceiveProps: function(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this.getInitialState());
		}
	},
	render: function() {
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
					onClick={this.handleCheckboxClick}>
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
				onHide={this.handleCancel}
				ref="modal">
				<Modal.Header closeButton>
					<Modal.Title>Delete Account</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				{warning}
				<Form horizontal onSubmit={this.handleSubmit}>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Delete Account</Col>
						<Col xs={10}>
						<AccountCombobox
							includeRoot={false}
							accounts={this.props.accounts}
							accountChildren={this.props.accountChildren}
							value={this.state.accountid}
							onChange={this.handleChange}/>
						</Col>
					</FormGroup>
					{checkbox}
				</Form>
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup className="pull-right">
						<Button onClick={this.handleCancel} bsStyle="warning">Cancel</Button>
						<Button onClick={this.handleSubmit} bsStyle="success">Delete Account</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
});

const AccountTreeNode = React.createClass({
	getInitialState: function() {
		return {expanded: false};
	},
	handleToggle: function(e) {
		e.preventDefault();
		this.setState({expanded:!this.state.expanded});
	},
	handleChildSelect: function(account) {
		if (this.props.onSelect != null)
			this.props.onSelect(account);
	},
	handleSelect: function() {
		if (this.props.onSelect != null)
			this.props.onSelect(this.props.account);
	},
	render: function() {
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
					onSelect={self.handleChildSelect}/>
		   );
		});
		var accounttreeClasses = "accounttree"
		var expandButton = [];
		if (children.length > 0) {
			expandButton.push((
				<Button onClick={this.handleToggle}
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
				<Button onClick={this.handleSelect}
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
});

const AccountTree = React.createClass({
	getInitialState: function() {
		return {height: 0};
	},
	handleSelect: function(account) {
		if (this.props.onSelect != null) {
			this.props.onSelect(account);
		}
	},
	resize: function() {
		var div = ReactDOM.findDOMNode(this);
		this.setState({height: div.parentElement.clientHeight - 73});
	},
	componentDidMount: function() {
		this.resize();
		var self = this;
		$(window).resize(function() {self.resize();});
	},
	render: function() {
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
					onSelect={this.handleSelect}/>));
			}
		}

		var style = {height: this.state.height + "px"};

		return (
			<div className="accounttree-root" style={style} >
				{children}
			</div>
		);
	}
});

module.exports = React.createClass({
	displayName: "AccountsTab",
	getInitialState: function() {
		return {
			creatingNewAccount: false,
			editingAccount: false,
			deletingAccount: false
		};
	},
	handleNewAccount: function() {
		this.setState({creatingNewAccount: true});
	},
	handleEditAccount: function() {
		this.setState({editingAccount: true});
	},
	handleDeleteAccount: function() {
		this.setState({deletingAccount: true});
	},
	handleCreationCancel: function() {
		this.setState({creatingNewAccount: false});
	},
	handleEditingCancel: function() {
		this.setState({editingAccount: false});
	},
	handleDeletionCancel: function() {
		this.setState({deletingAccount: false});
	},
	handleCreateAccount: function(account) {
		if (this.props.onCreateAccount != null)
			this.props.onCreateAccount(account);
		this.setState({creatingNewAccount: false});
	},
	handleUpdateAccount: function(account) {
		if (this.props.onUpdateAccount != null)
			this.props.onUpdateAccount(account);
		this.setState({editingAccount: false});
	},
	handleRemoveAccount: function(account) {
		if (this.props.onDeleteAccount != null)
			this.props.onDeleteAccount(account);
		this.setState({deletingAccount: false});
	},
	handleAccountSelected: function(account) {
		this.props.onSelectAccount(account.AccountId);
		this.props.onFetchTransactionPage(account, 20, 0);
	},
	render: function() {
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
						onCancel={this.handleCreationCancel}
						onSubmit={this.handleCreateAccount}
						security_list={this.props.security_list}/>
					<AddEditAccountModal
						show={this.state.editingAccount}
						editAccount={selectedAccount}
						accounts={this.props.accounts}
						accountChildren={this.props.accountChildren}
						onCancel={this.handleEditingCancel}
						onSubmit={this.handleUpdateAccount}
						security_list={this.props.security_list}/>
					<DeleteAccountModal
						show={this.state.deletingAccount}
						initialAccount={selectedAccount}
						accounts={this.props.accounts}
						accountChildren={this.props.accountChildren}
						onCancel={this.handleDeletionCancel}
						onSubmit={this.handleRemoveAccount}/>
					<AccountTree
						accounts={this.props.accounts}
						accountChildren={this.props.accountChildren}
						selectedAccount={this.props.selectedAccount}
						onSelect={this.handleAccountSelected}/>
					<ButtonGroup className="account-buttongroup">
						<Button onClick={this.handleNewAccount} bsStyle="success">
							<Glyphicon glyph='plus-sign' /></Button>
						<Button onClick={this.handleEditAccount}
								bsStyle="primary" disabled={disabled}>
							<Glyphicon glyph='cog' /></Button>
						<Button onClick={this.handleDeleteAccount}
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
						onCreateTransaction={this.props.onCreateTransaction}
						onUpdateTransaction={this.props.onUpdateTransaction}
						onDeleteTransaction={this.props.onDeleteTransaction}
						onSelectTransaction={this.props.onSelectTransaction}
						onUnselectTransaction={this.props.onUnselectTransaction}
						onFetchTransactionPage={this.props.onFetchTransactionPage}/>
				</Col>
			</Row></Grid>
		);
	}
});
