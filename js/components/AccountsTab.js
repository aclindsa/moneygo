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

var AccountTree = require('./AccountTree');
var AccountCombobox = require('./AccountCombobox');
var AccountRegister = require('./AccountRegister');
var AddEditAccountModal = require('./AddEditAccountModal');

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
