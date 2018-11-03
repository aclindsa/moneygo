var React = require('react');

var ReactBootstrap = require('react-bootstrap');
var Button = ReactBootstrap.Button;
var Collapse = ReactBootstrap.Collapse;
var Glyphicon = ReactBootstrap.Glyphicon;

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
						active={active}
						bsStyle="link"
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
		this.onSelect = this.handleSelect.bind(this);
	}
	handleSelect(account) {
		if (this.props.onSelectAccount != null) {
			this.props.onSelectAccount(account);
		}
		if (this.props.onSelect != null && this.props.onSelectKey != null) {
			this.props.onSelect(this.props.onSelectKey);
		}
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

		return (
			<div className="accounttree-root">
				{children}
			</div>
		);
	}
}

module.exports = AccountTree;
