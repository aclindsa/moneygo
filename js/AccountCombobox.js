var React = require('react');

var Combobox = require('react-widgets').Combobox;

var getAccountDisplayList = require('./utils.js').getAccountDisplayList;

module.exports = React.createClass({
	displayName: "AccountCombobox",
	getDefaultProps: function() {
		return {
			includeRoot: true,
			rootName: "New Top-level Account"
		};
	},
	handleAccountChange: function(account) {
		if (this.props.onChange != null &&
				account.hasOwnProperty('AccountId') &&
				(this.props.account_map.hasOwnProperty([account.AccountId]) ||
				 account.AccountId == -1)) {
			this.props.onChange(account)
		}
	},
	render: function() {
		var accounts = getAccountDisplayList(this.props.accounts, this.props.includeRoot, this.props.rootName);
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
				className={className} />
	   );
	}
});
