var React = require('react');

var ReactBootstrap = require('react-bootstrap');
var Jumbotron = ReactBootstrap.Jumbotron;
var Tabs = ReactBootstrap.Tabs;
var Tab = ReactBootstrap.Tab;
var Modal = ReactBootstrap.Modal;

var TopBar = require('./TopBar.js');
var NewUserForm = require('./NewUserForm.js');
var AccountSettingsModal = require('./AccountSettingsModal.js');
var AccountsTab = require('./AccountsTab.js');

module.exports = React.createClass({
	displayName: "MoneyGoApp",
	getInitialState: function() {
		return {
			hash: "home",
			session: new Session(),
			user: new User(),
			accounts: [],
			account_map: {},
			securities: [],
			security_map: {},
			error: new Error(),
			showAccountSettingsModal: false
		};
	},
	componentDidMount: function() {
		this.getSession();
		this.handleHashChange();
		if ("onhashchange" in window) {
			window.onhashchange = this.handleHashChange;
		}
	},
	handleHashChange: function() {
		var hash = location.hash.replace(/^#/, '');
		if (hash.length == 0)
			hash = "home";
		if (hash != this.state.hash)
			this.setHash(hash);
	},
	setHash: function(hash) {
		location.hash = hash;
		if (this.state.hash != hash)
		this.setState({hash: hash});
	},
	ajaxError: function(jqXHR, status, error) {
		var e = new Error();
		e.ErrorId = 5;
		e.ErrorString = "Request Failed: " + status + error;
		this.setState({error: e});
	},
	getSession: function() {
		$.ajax({
			type: "GET",
			dataType: "json",
			url: "session/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				var s = new Session();
				e.fromJSON(data);
				if (e.isError()) {
					if (e.ErrorId != 1 /* Not Signed In*/)
						this.setState({error: e});
				} else {
					s.fromJSON(data);
				}
				this.setState({session: s});
				this.getUser();
				this.getAccounts();
				this.getSecurities();
			}.bind(this),
			error: this.ajaxError
		});
	},
	getUser: function() {
		if (!this.state.session.isSession())
			return;
		$.ajax({
			type: "GET",
			dataType: "json",
			url: "user/"+this.state.session.UserId+"/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				var u = new User();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					u.fromJSON(data);
				}
				this.setState({user: u});
			}.bind(this),
			error: this.ajaxError
		});
	},
	getSecurities: function() {
		if (!this.state.session.isSession()) {
			this.setState({securities: [], security_map: {}});
			return;
		}
		$.ajax({
			type: "GET",
			dataType: "json",
			url: "security/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				var securities = [];
				var security_map = {};
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					for (var i = 0; i < data.securities.length; i++) {
						var s = new Security();
						s.fromJSON(data.securities[i]);
						securities.push(s);
						security_map[s.SecurityId] = s;
					}
				}
				this.setState({securities: securities, security_map: security_map});
			}.bind(this),
			error: this.ajaxError
		});
	},
	getAccounts: function() {
		if (!this.state.session.isSession()) {
			this.setState({accounts: [], account_map: {}});
			return;
		}
		$.ajax({
			type: "GET",
			dataType: "json",
			url: "account/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				var accounts = [];
				var account_map = {};
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					for (var i = 0; i < data.accounts.length; i++) {
						var a = new Account();
						a.fromJSON(data.accounts[i]);
						accounts.push(a);
						account_map[a.AccountId] = a;
					}
					//Populate Children arrays in account objects
					for (var i = 0; i < accounts.length; i++) {
						var a = accounts[i];
						if (!a.isRootAccount())
							account_map[a.ParentAccountId].Children.push(a);
					}
				}
				this.setState({accounts: accounts, account_map: account_map});
			}.bind(this),
			error: this.ajaxError
		});
	},
	handleErrorClear: function() {
		this.setState({error: new Error()});
	},
	handleLoginSubmit: function(user) {
		$.ajax({
			type: "POST",
			dataType: "json",
			url: "session/",
			data: {user: user.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					this.getSession();
					this.setHash("home");
				}
			}.bind(this),
			error: this.ajaxError
		});
	},
	handleLogoutSubmit: function() {
		this.setState({accounts: [], account_map: {}});
		$.ajax({
			type: "DELETE",
			dataType: "json",
			url: "session/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				}
				this.setState({session: new Session(), user: new User()});
			}.bind(this),
			error: this.ajaxError
		});
	},
	handleAccountSettings: function() {
		this.setState({showAccountSettingsModal: true});
	},
	handleSettingsSubmitted: function(user) {
		this.setState({
			user: user,
			showAccountSettingsModal: false
		});
	},
	handleSettingsCanceled: function(user) {
		this.setState({showAccountSettingsModal: false});
	},
	handleCreateNewUser: function() {
		this.setHash("new_user");
	},
	handleGoHome: function(user) {
		this.setHash("home");
	},
	handleCreateAccount: function(account) {
		$.ajax({
			type: "POST",
			dataType: "json",
			url: "account/",
			data: {account: account.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					this.getAccounts();
				}
			}.bind(this),
			error: this.ajaxError
		});
	},
	handleUpdateAccount: function(account) {
		$.ajax({
			type: "PUT",
			dataType: "json",
			url: "account/"+account.AccountId+"/",
			data: {account: account.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					this.getAccounts();
				}
			}.bind(this),
			error: this.ajaxError
		});
	},
	handleDeleteAccount: function(account) {
		$.ajax({
			type: "DELETE",
			dataType: "json",
			url: "account/"+account.AccountId+"/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					this.getAccounts();
				}
			}.bind(this),
			error: this.ajaxError
		});
	},
	render: function() {
		var mainContent;
		if (this.state.hash == "new_user") {
			mainContent = <NewUserForm onNewUser={this.handleGoHome} onCancel={this.handleGoHome}/>
		} else {
			if (this.state.user.isUser())
				mainContent = (
					<Tabs defaultActiveKey={1}>
						<Tab title="Accounts" eventKey={1} >
						<AccountsTab
							className="fullheight"
							accounts={this.state.accounts}
							account_map={this.state.account_map}
							securities={this.state.securities}
							security_map={this.state.security_map}
							onCreateAccount={this.handleCreateAccount}
							onUpdateAccount={this.handleUpdateAccount}
							onDeleteAccount={this.handleDeleteAccount} />
						</Tab>
						<Tab title="Scheduled Transactions" eventKey={2} >Scheduled transactions go here...</Tab>
						<Tab title="Budgets" eventKey={3} >Budgets go here...</Tab>
						<Tab title="Reports" eventKey={4} >Reports go here...</Tab>
					</Tabs>);
			else
				mainContent = (
					<Jumbotron>
						<center>
							<h1>Money<i>Go</i></h1>
							<p><i>Go</i> manage your money.</p>
						</center>
					</Jumbotron>);
		}

		return (
			<div className="fullheight ui">
				<TopBar
					error={this.state.error}
					onErrorClear={this.handleErrorClear}
					onLoginSubmit={this.handleLoginSubmit}
					onCreateNewUser={this.handleCreateNewUser}
					user={this.state.user}
					onAccountSettings={this.handleAccountSettings}
					onLogoutSubmit={this.handleLogoutSubmit} />
				{mainContent}
				<AccountSettingsModal
					show={this.state.showAccountSettingsModal}
					user={this.state.user}
					onSubmit={this.handleSettingsSubmitted}
					onCancel={this.handleSettingsCanceled}/>
			</div>
		);
	}
});
