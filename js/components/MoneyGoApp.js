var React = require('react');

var ReactBootstrap = require('react-bootstrap');
var Jumbotron = ReactBootstrap.Jumbotron;
var Tabs = ReactBootstrap.Tabs;
var Tab = ReactBootstrap.Tab;
var Modal = ReactBootstrap.Modal;

var TopBarContainer = require('../containers/TopBarContainer');
var NewUserForm = require('./NewUserForm');
var AccountSettingsModalContainer = require('../containers/AccountSettingsModalContainer');
var AccountsTabContainer = require('../containers/AccountsTabContainer');
var SecuritiesTabContainer = require('../containers/SecuritiesTabContainer');

module.exports = React.createClass({
	displayName: "MoneyGoApp",
	getInitialState: function() {
		return {
			hash: "home",
			showAccountSettingsModal: false
		};
	},
	componentDidMount: function() {
		this.props.tryResumingSession();
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
	handleAccountSettings: function() {
		this.setState({showAccountSettingsModal: true});
	},
	handleSettingsSubmitted: function(user) {
		this.setState({
			showAccountSettingsModal: false
		});
	},
	handleSettingsCanceled: function() {
		this.setState({showAccountSettingsModal: false});
	},
	handleCreateNewUser: function() {
		this.setHash("new_user");
	},
	handleGoHome: function() {
		this.setHash("home");
	},
	render: function() {
		var mainContent;
		if (this.state.hash == "new_user") {
			mainContent = <NewUserForm onNewUser={this.handleGoHome} onCancel={this.handleGoHome}/>
		} else {
			if (this.props.user.isUser())
				mainContent = (
					<Tabs defaultActiveKey={1} id='mainNavigationTabs'>
						<Tab title="Accounts" eventKey={1} >
						<AccountsTabContainer
							className="fullheight" />
						</Tab>
						<Tab title="Securities" eventKey={2} >
						<SecuritiesTabContainer
							className="fullheight" />
						</Tab>
						<Tab title="Scheduled Transactions" eventKey={3} >Scheduled transactions go here...</Tab>
						<Tab title="Budgets" eventKey={4} >Budgets go here...</Tab>
						<Tab title="Reports" eventKey={5} >Reports go here...</Tab>
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
				<TopBarContainer
					onCreateNewUser={this.handleCreateNewUser}
					onAccountSettings={this.handleAccountSettings} />
				{mainContent}
				<AccountSettingsModalContainer
					show={this.state.showAccountSettingsModal}
					onSubmit={this.handleSettingsSubmitted}
					onCancel={this.handleSettingsCanceled}/>
			</div>
		);
	}
});
