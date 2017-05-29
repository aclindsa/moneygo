var React = require('react');

var Combobox = require('react-widgets').Combobox;

var getAccountDisplayList = require('../utils').getAccountDisplayList;

module.exports = React.createClass({
	displayName: "AccountCombobox",
	getDefaultProps: function() {
		return {
			includeRoot: true,
			disabled: false,
			rootName: "New Top-level Account"
		};
	},
	handleAccountChange: function(account) {
		if (this.props.onChange != null &&
				account.hasOwnProperty('AccountId') &&
				(this.props.accounts.hasOwnProperty([account.AccountId]) ||
				 account.AccountId == -1)) {
			this.props.onChange(account)
		}
	},
	render: function() {
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
				onChange={this.handleAccountChange}
				ref="account"
				disabled={this.props.disabled}
				suggest
				filter='contains'
				className={className} />
	   );
	}
});
