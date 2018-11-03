var React = require('react');
var ReactDOM = require('react-dom');

var ReactBootstrap = require('react-bootstrap');
var Jumbotron = ReactBootstrap.Jumbotron;
var Tabs = ReactBootstrap.Tabs;
var Tab = ReactBootstrap.Tab;
var Nav = ReactBootstrap.Nav;
var NavDropdown = ReactBootstrap.NavDropdown;
var NavItem = ReactBootstrap.NavItem;
var Col = ReactBootstrap.Col;
var Row = ReactBootstrap.Row;
var Modal = ReactBootstrap.Modal;
var FormControl = ReactBootstrap.FormControl;
var InputGroup = ReactBootstrap.InputGroup;
var Button = ReactBootstrap.Button;
var Glyphicon = ReactBootstrap.Glyphicon;

var MenuItem = ReactBootstrap.MenuItem;
var SplitButton = ReactBootstrap.SplitButton;

var TopBarContainer = require('../containers/TopBarContainer');
var NewUserModalContainer = require('../containers/NewUserModalContainer');
var AccountSettingsModalContainer = require('../containers/AccountSettingsModalContainer');
var AccountsTabContainer = require('../containers/AccountsTabContainer');
var SecuritiesTabContainer = require('../containers/SecuritiesTabContainer');
var ReportsTabContainer = require('../containers/ReportsTabContainer');
var AddEditAccountModal = require('./AddEditAccountModal');

var AccountTree = require('./AccountTree');

var utils = require('../utils');

class MoneyGoApp extends React.Component {
	constructor() {
		super();
		this.state = {
			tab: 1,
			accountFilter: "",
			showNewUserModal: false,
			showSettingsModal: false,
			creatingNewAccount: false
		};
		this.onShowSettings = this.handleShowSettings.bind(this);
		this.onHideSettings = this.handleHideSettings.bind(this);
		this.onShowNewUser = this.handleShowNewUser.bind(this);
		this.onHideNewUser = this.handleHideNewUser.bind(this);
		this.onSelectTab = this.handleSelectTab.bind(this);
		this.onAccountSelected = this.handleAccountSelected.bind(this);
		this.onNewAccount = this.handleNewAccount.bind(this);
		this.onNewAccountCancel = this.handleNewAccountCancel.bind(this);
		this.onCreateAccount = this.handleCreateAccount.bind(this);
		this.onAccountFilterChange = this.handleAccountFilterChange.bind(this);
		this.onClearAccountFilter = this.handleClearAccountFilter.bind(this);
	}
	componentDidMount() {
		this.props.tryResumingSession();
		this.props.fetchCurrencies();
	}
	handleShowSettings() {
		this.setState({showSettingsModal: true});
	}
	handleHideSettings(user) {
		this.setState({showSettingsModal: false});
	}
	handleShowNewUser() {
		this.setState({showNewUserModal: true});
	}
	handleHideNewUser() {
		this.setState({showNewUserModal: false});
	}
	handleSelectTab(key) {
		console.log(key);
		if (key != undefined) {
			this.setState({tab: key});
		}
	}
	handleAccountSelected(account) {
		this.setState({
			tab: 1,
			accountFilter: ""});
		this.props.onSelectAccount(account.AccountId);
		this.props.onFetchTransactionPage(account, 20, 0);
	}
	handleNewAccount() {
		this.setState({creatingNewAccount: true});
	}
	handleNewAccountCancel() {
		this.setState({creatingNewAccount: false});
	}
	handleCreateAccount(account) {
		this.props.onCreateAccount(account);
		this.setState({creatingNewAccount: false});
	}
	handleAccountFilterChange() {
		this.setState({accountFilter: ReactDOM.findDOMNode(this.refs.accountFilter).value});
	}
	handleClearAccountFilter() {
		this.setState({accountFilter: ""});
	}
	render() {
		var mainContent;
		if (this.props.user.isUser()) {
			var accountElements = [];
			if (this.state.accountFilter.length > 0) {
				var filterRegex = new RegExp(this.state.accountFilter, "i")
				for (var accountId in this.props.accounts) {
					var account = this.props.accounts[accountId];
					var fullName = utils.getAccountDisplayName(account, this.props.accounts);
					if (fullName.match(filterRegex)) {
						var clojure = function(self, account) {
							return function() {
								self.handleAccountSelected(account);
							};
						}(this, account);
						accountElements.push((<MenuItem key={accountId} onSelect={clojure}>{fullName}</MenuItem>));
					}
				}
			} else {
				accountElements.push((<MenuItem key={1} divider />));
				accountElements.push((<AccountTree key={2}
						accounts={this.props.accounts}
						accountChildren={this.props.accountChildren}
						selectedAccount={this.props.selectedAccount}
						onSelectAccount={this.onAccountSelected}
						onSelectKey={1} />));
			}

			mainContent = (
				<Tab.Container id="main-ui-navigation" activeKey={this.state.tab} onSelect={this.onSelectTab}>
					<Row className="clearfix">
						<Col sm={12}>
							<Nav bsStyle="tabs">
								<NavDropdown eventKey={1} title="Accounts">
									<MenuItem onSelect={this.onNewAccount}>New Account</MenuItem>
									<MenuItem divider />
									<MenuItem disabled className="account-filter-menuitem">
										<InputGroup>
											<FormControl
												onKeyDown={function(e){if (e.key == " ") e.stopPropagation();}}
												type="text"
												placeholder="Search..."
												value={this.state.accountFilter}
												onChange={this.onAccountFilterChange}
												ref="accountFilter" />
											<InputGroup.Button>
												<Button className="clear-account-filter" onClick={this.onClearAccountFilter}>
													<Glyphicon glyph="remove"/>
												</Button>
											</InputGroup.Button>
										</InputGroup>
									</MenuItem>
									{accountElements}
								</NavDropdown>
								<NavItem eventKey={2}>Securities</NavItem>
								<NavItem eventKey={3}>Scheduled Transactions</NavItem>
								<NavItem eventKey={4}>Budgets</NavItem>
								<NavItem eventKey={5}>Reports</NavItem>
							</Nav>
						</Col>
						<Col sm={12}>
							<Tab.Content>
								<Tab.Pane eventKey={1}>
									<AccountsTabContainer/>
								</Tab.Pane>
								<Tab.Pane eventKey={2}>
									<SecuritiesTabContainer/>
								</Tab.Pane>
								<Tab.Pane eventKey={3}>
									Scheduled transactions go here...
								</Tab.Pane>
								<Tab.Pane eventKey={4}>
									Budgets go here...
								</Tab.Pane>
								<Tab.Pane eventKey={5}>
									<ReportsTabContainer/>
								</Tab.Pane>
							</Tab.Content>
						</Col>
					</Row>
				</Tab.Container>);
		} else {
			mainContent = (
				<Jumbotron>
					<center>
						<h1>Money<i>Go</i></h1>
						<p><i>Go</i> manage your money.</p>
					</center>
				</Jumbotron>);
		}

		return (
			<div className="ui">
				<TopBarContainer
					onCreateNewUser={this.onShowNewUser}
					onSettings={this.onShowSettings} />
				{mainContent}
				<NewUserModalContainer
					show={this.state.showNewUserModal}
					onSubmit={this.onHideNewUser}
					onCancel={this.onHideNewUser}/>
				<AccountSettingsModalContainer
					show={this.state.showSettingsModal}
					onSubmit={this.onHideSettings}
					onCancel={this.onHideSettings}/>
				<AddEditAccountModal
					show={this.state.creatingNewAccount}
					initialParentAccount={this.props.selectedAccount}
					accounts={this.props.accounts}
					accountChildren={this.props.accountChildren}
					onCancel={this.onNewAccountCancel}
					onSubmit={this.onCreateAccount}
					security_list={this.props.security_list}/>
			</div>
		);
	}
}

module.exports = MoneyGoApp;
