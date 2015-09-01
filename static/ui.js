// Import all the objects we want to use from ReactBootstrap
var Jumbotron = ReactBootstrap.Jumbotron;
var TabbedArea = ReactBootstrap.TabbedArea;
var TabPane = ReactBootstrap.TabPane;
var Panel = ReactBootstrap.Panel;
var ButtonGroup = ReactBootstrap.ButtonGroup;
var Modal = ReactBootstrap.Modal;

const NewUserForm = React.createClass({
	getInitialState: function() {
		return {error: "",
			name: "",
			username: "",
			email: "",
			password: "",
			confirm_password: "",
			passwordChanged: false,
			initial_password: ""};
	},
	passwordValidationState: function() {
		if (this.state.passwordChanged) {
			if (this.state.password.length >= 10)
				return "success";
			else if (this.state.password.length >= 6)
				return "warning";
			else
				return "error";
		}
	},
	confirmPasswordValidationState: function() {
		if (this.state.confirm_password.length > 0) {
			if (this.state.confirm_password == this.state.password)
				return "success";
			else
				return "error";
		}
	},
	handleCancel: function() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	},
	handleChange: function() {
		if (this.refs.password.getValue() != this.state.initial_password)
			this.setState({passwordChanged: true});
		this.setState({
			name: this.refs.name.getValue(),
			username: this.refs.username.getValue(),
			email: this.refs.email.getValue(),
			password: this.refs.password.getValue(),
			confirm_password: this.refs.confirm_password.getValue()
		});
	},
	handleSubmit: function(e) {
		var u = new User();
		var error = "";
		e.preventDefault();

		u.Name = this.state.name;
		u.Username = this.state.username;
		u.Email = this.state.email;
		u.Password = this.state.password;
		if (u.Password != this.state.confirm_password) {
			this.setState({error: "Error: password do not match"});
			return;
		}

		this.handleCreateNewUser(u);
	},
	handleCreateNewUser: function(user) {
		$.ajax({
			type: "POST",
			dataType: "json",
			url: "user/",
			data: {user: user.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					this.props.onNewUser();
				}
			}.bind(this),
			error: function(jqXHR, status, error) {
				var e = new Error();
				e.ErrorId = 5;
				e.ErrorString = "Request Failed: " + status + error;
				this.setState({error: e});
			}.bind(this),
		});
	},
	render: function() {
		var title = <h3>Create New User</h3>;
		return (
			<Panel header={title} bsStyle="info">
				<span color="red">{this.state.error}</span>
				<form onSubmit={this.handleSubmit}
						className="form-horizontal">
					<Input type="text"
							label="Name"
							value={this.state.name}
							onChange={this.handleChange}
							ref="name"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"/>
					<Input type="text"
							label="Username"
							value={this.state.username}
							onChange={this.handleChange}
							ref="username"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"/>
					<Input type="email"
							label="Email"
							value={this.state.email}
							onChange={this.handleChange}
							ref="email"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"/>
					<Input type="password"
							label="Password"
							value={this.state.password}
							onChange={this.handleChange}
							ref="password"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"
							bsStyle={this.passwordValidationState()}
							hasFeedback/>
					<Input type="password"
							label="Confirm Password"
							value={this.state.confirm_password}
							onChange={this.handleChange}
							ref="confirm_password"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"
							bsStyle={this.confirmPasswordValidationState()}
							hasFeedback/>
					<ButtonGroup className="pull-right">
						<Button onClick={this.handleCancel}
								bsStyle="warning">Cancel</Button>
						<Button type="submit"
								bsStyle="success">Create New User</Button>
					</ButtonGroup>
				</form>
			</Panel>
		);
	}
});

const AccountSettingsModal = React.createClass({
	_getInitialState: function(props) {
		return {error: "",
			name: props.user.Name,
			username: props.user.Username,
			email: props.user.Email,
			password: BogusPassword,
			confirm_password: BogusPassword,
			passwordChanged: false,
			initial_password: BogusPassword};
	},
	getInitialState: function() {
		 return this._getInitialState(this.props);
	},
	componentWillReceiveProps: function(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this._getInitialState(nextProps));
		}
	},
	passwordValidationState: function() {
		if (this.state.passwordChanged) {
			if (this.state.password.length >= 10)
				return "success";
			else if (this.state.password.length >= 6)
				return "warning";
			else
				return "error";
		}
	},
	confirmPasswordValidationState: function() {
		if (this.state.confirm_password.length > 0) {
			if (this.state.confirm_password == this.state.password)
				return "success";
			else
				return "error";
		}
	},
	handleCancel: function() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	},
	handleChange: function() {
		if (this.refs.password.getValue() != this.state.initial_password)
			this.setState({passwordChanged: true});
		this.setState({
			name: this.refs.name.getValue(),
			username: this.refs.username.getValue(),
			email: this.refs.email.getValue(),
			password: this.refs.password.getValue(),
			confirm_password: this.refs.confirm_password.getValue()
		});
	},
	handleSubmit: function(e) {
		var u = new User();
		var error = "";
		e.preventDefault();

		u.UserId = this.props.user.UserId;
		u.Name = this.state.name;
		u.Username = this.state.username;
		u.Email = this.state.email;
		if (this.state.passwordChanged) {
			u.Password = this.state.password;
			if (u.Password != this.state.confirm_password) {
				this.setState({error: "Error: password do not match"});
				return;
			}
		} else {
			u.Password = BogusPassword;
		}

		this.handleSaveSettings(u);
	},
	handleSaveSettings: function(user) {
		$.ajax({
			type: "PUT",
			dataType: "json",
			url: "user/"+user.UserId+"/",
			data: {user: user.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					user.Password = "";
					this.props.onSubmit(user);
				}
			}.bind(this),
			error: function(jqXHR, status, error) {
				var e = new Error();
				e.ErrorId = 5;
				e.ErrorString = "Request Failed: " + status + error;
				this.setState({error: e});
			}.bind(this),
		});
	},
	render: function() {
		return (
			<Modal show={this.props.show} onHide={this.handleCancel} bsSize="large">
				<Modal.Header closeButton>
					<Modal.Title>Edit Account Settings</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<span color="red">{this.state.error}</span>
				<form onSubmit={this.handleSubmit}
						className="form-horizontal">
					<Input type="text"
							label="Name"
							value={this.state.name}
							onChange={this.handleChange}
							ref="name"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"/>
					<Input type="text"
							label="Username"
							value={this.state.username}
							onChange={this.handleChange}
							ref="username"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"/>
					<Input type="email"
							label="Email"
							value={this.state.email}
							onChange={this.handleChange}
							ref="email"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"/>
					<Input type="password"
							label="Password"
							value={this.state.password}
							onChange={this.handleChange}
							ref="password"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"
							bsStyle={this.passwordValidationState()}
							hasFeedback/>
					<Input type="password"
							label="Confirm Password"
							value={this.state.confirm_password}
							onChange={this.handleChange}
							ref="confirm_password"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"
							bsStyle={this.confirmPasswordValidationState()}
							hasFeedback/>
				</form>
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup>
						<Button onClick={this.handleCancel} bsStyle="warning">Cancel</Button>
						<Button onClick={this.handleSubmit} bsStyle="success">Save Settings</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
});

const MoneyGoApp = React.createClass({
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
				mainContent =
					<TabbedArea defaultActiveKey='1' className="">
						<TabPane tab="Accounts" eventKey='1' className="fullheight">
						<AccountsTab
							className="fullheight"
							accounts={this.state.accounts}
							account_map={this.state.account_map}
							securities={this.state.securities}
							security_map={this.state.security_map}
							onCreateAccount={this.handleCreateAccount}
							onUpdateAccount={this.handleUpdateAccount}
							onDeleteAccount={this.handleDeleteAccount} />
						</TabPane>
						<TabPane tab="Scheduled Transactions" eventKey='2' className="fullheight">Scheduled transactions go here...</TabPane>
						<TabPane tab="Budgets" eventKey='3' className="fullheight">Budgets go here...</TabPane>
						<TabPane tab="Reports" eventKey='4' className="fullheight">Reports go here...</TabPane>
					</TabbedArea>
			else
				mainContent =
					<Jumbotron>
						<center>
							<h1>Money<i>Go</i></h1>
							<p><i>Go</i> manage your money.</p>
						</center>
					</Jumbotron>
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

React.render(<MoneyGoApp />, document.getElementById("content"));
