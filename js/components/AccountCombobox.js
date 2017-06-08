var React = require('react');

var Combobox = require('react-widgets').Combobox;

var getAccountDisplayList = require('../utils').getAccountDisplayList;

class AccountCombobox extends React.Component {
	static get defaultProps() {
		return {
			includeRoot: true,
			disabled: false,
			rootName: "New Top-level Account"
		}
	}
	constructor() {
		super();
		this.onAccountChange = this.handleAccountChange.bind(this);
	}
	handleAccountChange(account) {
		if (this.props.onChange != null &&
				account.hasOwnProperty('AccountId') &&
				(this.props.accounts.hasOwnProperty([account.AccountId]) ||
				 account.AccountId == -1)) {
			this.props.onChange(account)
		}
	}
	render() {
		var accounts = getAccountDisplayList(this.props.accounts, this.props.accountChildren, this.props.includeRoot, this.props.rootName);
		var className = "";
		if (this.props.className)
			className = this.props.className;
		return (
			<Combobox
				data={accounts}
				valueField='AccountId'
				textField='Name'
				defaultValue={this.props.value}
				onChange={this.onAccountChange}
				ref="account"
				disabled={this.props.disabled}
				suggest
				filter='contains'
				className={className} />
	   );
	}
}

module.exports = AccountCombobox;
