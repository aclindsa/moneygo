var connect = require('react-redux').connect;

var AccountActions = require('../actions/AccountActions');
var AccountsTab = require('../components/AccountsTab');

function mapStateToProps(state) {
	var security_list = [];
	for (var securityId in state.securities) {
		if (state.securities.hasOwnProperty(securityId))
			security_list.push(state.securities[securityId]);
	}
	return {
		accounts: state.accounts.map,
		accountChildren: state.accounts.children,
		securities: state.securities,
		security_list: security_list,
		selectedAccount: state.selectedAccount
	}
}

function mapDispatchToProps(dispatch) {
	return {
		onCreateAccount: function(account) {dispatch(AccountActions.create(account))},
		onUpdateAccount: function(account) {dispatch(AccountActions.update(account))},
		onDeleteAccount: function(accountId) {dispatch(AccountActions.remove(accountId))},
		onSelectAccount: function(accountId) {dispatch(AccountActions.select(accountId))}
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(AccountsTab)
