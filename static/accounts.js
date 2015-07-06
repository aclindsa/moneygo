// Import all the objects we want to use from ReactBootstrap
var ListGroup = ReactBootstrap.ListGroup;
var ListGroupItem = ReactBootstrap.ListGroupItem;

var Grid = ReactBootstrap.Grid;
var Row = ReactBootstrap.Row;
var Col = ReactBootstrap.Col;

var Button = ReactBootstrap.Button;
var ButtonGroup = ReactBootstrap.ButtonGroup;
var Glyphicon = ReactBootstrap.Glyphicon;

var Modal = ReactBootstrap.Modal;

var CollapsibleMixin = ReactBootstrap.CollapsibleMixin;

var Combobox = ReactWidgets.Combobox;

const recursiveAccountDisplayInfo = function(account, prefix) {
	var name = prefix + account.Name;
	var accounts = [{AccountId: account.AccountId, Name: name}];
	for (var i = 0; i < account.Children.length; i++)
		accounts = accounts.concat(recursiveAccountDisplayInfo(account.Children[i], name + "/"));
	return accounts
};
const getAccountDisplayList = function(account_list, includeRoot, rootName) {
	var accounts = []
	if (includeRoot)
		accounts.push({AccountId: -1, Name: rootName});
	for (var i = 0; i < account_list.length; i++) {
		if (account_list[i].isRootAccount())
			accounts = accounts.concat(recursiveAccountDisplayInfo(account_list[i], ""));
	}
	return accounts;
};

const AccountCombobox = React.createClass({
	getDefaultProps: function() {
		return {
			includeRoot: true,
			rootName: "New Top-level Account"
		};
	},
	handleAccountChange: function(account) {
		if (this.props.onSelect != null &&
				account.hasOwnProperty('AccountId') &&
				(this.props.account_map.hasOwnProperty([account.AccountId]) ||
				 account.AccountId == -1)) {
			this.props.onSelect(account)
		}
	},
	render: function() {
		var accounts = getAccountDisplayList(this.props.accounts, this.props.includeRoot, this.props.rootName);
		return (
			<Combobox
				data={accounts}
				valueField='AccountId'
				textField='Name'
				value={this.props.value}
				onSelect={this.handleAccountChange}
				ref="account" />
	   );
	}
});

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
			name: this.refs.name.getValue(),
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
				<form onSubmit={this.handleSubmit}
						className="form-horizontal">
					<Input type="text"
						label="Name"
						value={this.state.name}
						onChange={this.handleChange}
						ref="name"
						labelClassName="col-xs-2"
						wrapperClassName="col-xs-10"/>
					<Input wrapperClassName="wrapper"
						label="Parent Account"
						labelClassName="col-xs-2"
						wrapperClassName="col-xs-10">
					<AccountCombobox
						accounts={this.props.accounts}
						account_map={this.props.account_map}
						value={this.state.parentaccountid}
						rootName={rootName}
						onSelect={this.handleParentChange}
						ref="parent" />
					</Input>
					<Input wrapperClassName="wrapper"
						label="Security"
						labelClassName="col-xs-2"
						wrapperClassName="col-xs-10">
					<Combobox
						data={this.props.securities}
						valueField='SecurityId'
						textField='Name'
						value={this.state.security}
						onSelect={this.handleSecurityChange}
						ref="security" />
					</Input>
					<Input wrapperClassName="wrapper"
						label="Account Type"
						labelClassName="col-xs-2"
						wrapperClassName="col-xs-10">
					<Combobox
						data={AccountTypeList}
						valueField='TypeId'
						textField='Name'
						value={this.state.type}
						onSelect={this.handleTypeChange}
						ref="type" />
					</Input>
				</form>
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
		if (this.props.account_map.hasOwnProperty(this.state.accountid)) {
			if (this.state.checked) {
				if (this.props.onSubmit != null)
					this.props.onSubmit(this.props.account_map[this.state.accountid]);
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
		if (this.props.account_map.hasOwnProperty(this.state.accountid)) {
			var parentAccountId = this.props.account_map[this.state.accountid].ParentAccountId;
			var parentAccount = "will be deleted and any child accounts will become top-level accounts.";
			if (parentAccountId != -1)
				parentAccount = "and any child accounts will be re-parented to: " + this.props.account_map[parentAccountId].Name;

			var warningString = "I understand that deleting this account cannot be undone and that all transactions " + parentAccount;
			checkbox = (<Input
				type='checkbox'
				checked={this.state.checked ? "checked" : ""}
				onClick={this.handleCheckboxClick}
				label={warningString}
				wrapperClassName="col-xs-offset-2 col-xs-10"/>);
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
				<form onSubmit={this.handleSubmit}
						className="form-horizontal">
					<Input wrapperClassName="wrapper"
						label="Delete Account"
						labelClassName="col-xs-2"
						wrapperClassName="col-xs-10">
					<AccountCombobox
						includeRoot={false}
						accounts={this.props.accounts}
						account_map={this.props.account_map}
						value={this.state.accountid}
						onSelect={this.handleChange}/>
					</Input>
					{checkbox}
				</form>
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
	mixins: [CollapsibleMixin],
	getCollapsibleDOMNode: function() {
		return React.findDOMNode(this.refs.children);
	},
	getCollapsibleDimensionValue: function() {
		return React.findDOMNode(this.refs.children).scrollHeight;
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
		var styles = this.getCollapsibleClassSet();
		var glyph = this.isExpanded() ? 'minus' : 'plus';
		var active = (this.props.selectedAccount != null &&
			this.props.account.AccountId == this.props.selectedAccount.AccountId);
		var buttonStyle = active ? "info" : "link";

		var self = this;
		var children = this.props.account.Children.map(function(account) {
			return (
				<AccountTreeNode
					account={account}
					selectedAccount={self.props.selectedAccount}
					onSelect={self.handleChildSelect}/>
		   );
		});
		var accounttreeClasses = "accounttree"
		var expandButton = [];
		if (children.length > 0)
			expandButton.push((
				<Button onClick={this.handleToggle}
						bsSize="xsmall"
						bsStyle="link"
						className="accounttree-expandbutton">
					<Glyphicon glyph={glyph} bsSize="xsmall"/>
				</Button>
			));
		else
			accounttreeClasses += "-nochildren";
		return (
			<div className={accounttreeClasses}>
				{expandButton}
				<Button onClick={this.handleSelect}
						bsStyle={buttonStyle}
						className="accounttree-name">
					{this.props.account.Name}
				</Button>
				<div ref='children' className={classNames(styles)}>
					{children}
				</div>
			</div>
		);
	}
});

const AccountTree = React.createClass({
	getInitialState: function() {
		return {selectedAccount: null};
	},
	handleSelect: function(account) {
		this.setState({selectedAccount: account});
		if (this.props.onSelect != null) {
			this.props.onSelect(account);
		}
	},
	render: function() {
		var accounts = this.props.accounts;

		var children = [];
		for (var i = 0; i < accounts.length; i++) {
			if (accounts[i].isRootAccount())
				children.push((<AccountTreeNode
					account={accounts[i]}
					selectedAccount={this.state.selectedAccount}
					onSelect={this.handleSelect}/>));
		}

		return (
			<div className="accounttree-root">
				{children}
			</div>
		);
	}
});

const AccountsTab = React.createClass({
	getInitialState: function() {
		return {
			selectedAccount: null,
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
		this.setState({deletingAccount: false,
			selectedAccount: null});
	},
	handleAccountSelected: function(account) {
		this.setState({selectedAccount: account});
	},
	render: function() {
		var accounts = this.props.accounts;
		var account_map = this.props.account_map;

		var disabled = (this.state.selectedAccount == null) ? "disabled" : "";

		return (
			<Grid fluid><Row>
				<Col xs={2}>
					<AddEditAccountModal
						show={this.state.creatingNewAccount}
						initialParentAccount={this.state.selectedAccount}
						accounts={this.props.accounts}
						account_map={this.props.account_map}
						onCancel={this.handleCreationCancel}
						onSubmit={this.handleCreateAccount}
						securities={this.props.securities}/>
					<AddEditAccountModal
						show={this.state.editingAccount}
						editAccount={this.state.selectedAccount}
						accounts={this.props.accounts}
						account_map={this.props.account_map}
						onCancel={this.handleEditingCancel}
						onSubmit={this.handleUpdateAccount}
						securities={this.props.securities}/>
					<DeleteAccountModal
						show={this.state.deletingAccount}
						initialAccount={this.state.selectedAccount}
						accounts={this.props.accounts}
						account_map={this.props.account_map}
						onCancel={this.handleDeletionCancel}
						onSubmit={this.handleRemoveAccount}/>
					<AccountTree
						accounts={accounts}
						onSelect={this.handleAccountSelected}/>
					<ButtonGroup className="pull-right">
						<Button onClick={this.handleNewAccount} bsStyle="success">
							<Glyphicon glyph='plus-sign' /></Button>
						<Button onClick={this.handleEditAccount}
								bsStyle="primary" disabled={disabled}>
							<Glyphicon glyph='cog' /></Button>
						<Button onClick={this.handleDeleteAccount}
								bsStyle="danger" disabled={disabled}>
							<Glyphicon glyph='trash' /></Button>
					</ButtonGroup>
				</Col><Col xs={10}>
					blah
				</Col>
			</Row></Grid>
		);
	}
});
