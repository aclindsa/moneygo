var connect = require('react-redux').connect;

var UserActions = require('../actions/UserActions');
var AccountActions = require('../actions/AccountActions');
var TransactionActions = require('../actions/TransactionActions');
var SecurityTemplateActions = require('../actions/SecurityTemplateActions');

var MoneyGoApp = require('../components/MoneyGoApp');

function mapStateToProps(state) {
	return {
		user: state.user,
		accounts: state.accounts.map,
		accountChildren: state.accounts.children,
		selectedAccount: state.selectedAccount,
		security_list: state.securities.list,
	}
}

function mapDispatchToProps(dispatch) {
	return {
		tryResumingSession: function() {dispatch(UserActions.tryResumingSession())},
		fetchCurrencies: function() {dispatch(SecurityTemplateActions.fetchCurrencies())},
		onCreateAccount: function(account) {dispatch(AccountActions.create(account))},
		onSelectAccount: function(accountId) {dispatch(AccountActions.select(accountId))},
		onFetchTransactionPage: function(account, pageSize, page) {dispatch(TransactionActions.fetchPage(account, pageSize, page))},
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(MoneyGoApp)
