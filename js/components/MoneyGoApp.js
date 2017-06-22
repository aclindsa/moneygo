var React = require('react');

var ReactBootstrap = require('react-bootstrap');
var Jumbotron = ReactBootstrap.Jumbotron;
var Tabs = ReactBootstrap.Tabs;
var Tab = ReactBootstrap.Tab;
var Modal = ReactBootstrap.Modal;

var TopBarContainer = require('../containers/TopBarContainer');
var NewUserModalContainer = require('../containers/NewUserModalContainer');
var AccountSettingsModalContainer = require('../containers/AccountSettingsModalContainer');
var AccountsTabContainer = require('../containers/AccountsTabContainer');
var SecuritiesTabContainer = require('../containers/SecuritiesTabContainer');
var ReportsTabContainer = require('../containers/ReportsTabContainer');

class MoneyGoApp extends React.Component {
	constructor() {
		super();
		this.state = {
			showNewUserModal: false,
			showAccountSettingsModal: false
		};
		this.onShowSettings = this.handleShowSettings.bind(this);
		this.onHideSettings = this.handleHideSettings.bind(this);
		this.onShowNewUser = this.handleShowNewUser.bind(this);
		this.onHideNewUser = this.handleHideNewUser.bind(this);
	}
	componentDidMount() {
		this.props.tryResumingSession();
		this.props.fetchCurrencies();
	}
	handleShowSettings() {
		this.setState({showAccountSettingsModal: true});
	}
	handleHideSettings(user) {
		this.setState({showAccountSettingsModal: false});
	}
	handleShowNewUser() {
		this.setState({showNewUserModal: true});
	}
	handleHideNewUser() {
		this.setState({showNewUserModal: false});
	}
	render() {
		var mainContent;
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
					<Tab title="Reports" eventKey={5} >
					<ReportsTabContainer
						className="fullheight" />
					</Tab>
				</Tabs>);
		else
			mainContent = (
				<Jumbotron>
					<center>
						<h1>Money<i>Go</i></h1>
						<p><i>Go</i> manage your money.</p>
					</center>
				</Jumbotron>);

		return (
			<div className="fullheight ui">
				<TopBarContainer
					onCreateNewUser={this.onShowNewUser}
					onAccountSettings={this.onShowSettings} />
				{mainContent}
				<NewUserModalContainer
					show={this.state.showNewUserModal}
					onSubmit={this.onHideNewUser}
					onCancel={this.onHideNewUser}/>
				<AccountSettingsModalContainer
					show={this.state.showAccountSettingsModal}
					onSubmit={this.onHideSettings}
					onCancel={this.onHideSettings}/>
			</div>
		);
	}
}

module.exports = MoneyGoApp;
