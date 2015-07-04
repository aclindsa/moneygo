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
	handleAccountChange: function(account) {
		if (this.props.onSelect != null &&
				account.hasOwnProperty('AccountId') &&
				this.props.account_map.hasOwnProperty([account.AccountId])) {
			this.props.onSelect(this.props.account_map[account.AccountId])
		}
	},
	render: function() {
		var accounts = getAccountDisplayList(this.props.accounts, true, "New Root Account");
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

const NewAccountModal = React.createClass({
	getInitialState: function() {
		return {
			security: 1,
			parentaccountid: -1,
			type: 1,
			name: ""
		};
	},
	handleCancel: function() {
		if (this.props.onCancel != null)
			this.props.onCancel();
		this.setState(this.getInitialState());
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

		a.Name = this.state.name;
		a.ParentAccountId = this.state.parentaccountid;
		a.SecurityId = this.state.security;
		a.Type = this.state.type;

		this.handleSaveSettings(a);
		this.setState(this.getInitialState());
	},
	handleSaveSettings: function(account) {
		if (this.props.onSubmit != null)
			this.props.onSubmit(account);
	},
	render: function() {
		return (
			<Modal show={this.props.show} onHide={this.handleCancel}>
				<Modal.Header closeButton>
					<Modal.Title>Create New Account</Modal.Title>
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
						<Button onClick={this.handleSubmit} bsStyle="success">Create Account</Button>
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
			creatingNewAccount: false
		};
	},
	handleNewAccount: function() {
		this.setState({creatingNewAccount: true});
	},
	handleEditAccount: function() {
		console.log("handleEditAccount");
	},
	handleDeleteAccount: function() {
		console.log("handleDeleteAccount");
	},
	handleCreationCancel: function() {
		this.setState({creatingNewAccount: false});
	},
	handleCreateAccount: function(account) {
		if (this.props.onCreateAccount != null)
			this.props.onCreateAccount(account);
		this.setState({creatingNewAccount: false});
	},
	handleAccountSelected: function(account) {
		this.setState({selectedAccount: account});
	},
	render: function() {
		var accounts = this.props.accounts;
		var account_map = this.props.account_map;

		return (
			<Grid fluid><Row>
				<Col xs={2}>
					<NewAccountModal
						show={this.state.creatingNewAccount}
						accounts={this.props.accounts}
						account_map={this.props.account_map}
						onCancel={this.handleCreationCancel}
						onSubmit={this.handleCreateAccount}
						securities={this.props.securities}/>
					<AccountTree
						accounts={accounts}
						onSelect={this.handleAccountSelected}/>
					<ButtonGroup className="pull-right">
						<Button onClick={this.handleNewAccount} bsStyle="success">
							<Glyphicon glyph='plus-sign' /></Button>
						<Button onClick={this.handleEditAccount} bsStyle="primary">
							<Glyphicon glyph='cog' /></Button>
						<Button onClick={this.handleDeleteAccount} bsStyle="danger">
							<Glyphicon glyph='trash' /></Button>
					</ButtonGroup>
				</Col><Col xs={10}>
					blah
				</Col>
			</Row></Grid>
		);
	}
});
