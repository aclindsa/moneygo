var connect = require('react-redux').connect;

var AccountActions = require('../actions/AccountActions');
var TransactionActions = require('../actions/TransactionActions');
var ImportActions = require('../actions/ImportActions');

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
		selectedAccount: state.selectedAccount,
		transactions: state.transactions,
		transactionPage: state.transactionPage,
		imports: state.imports
	}
}

function mapDispatchToProps(dispatch) {
	return {
		onFetchAllAccounts: function() {dispatch(AccountActions.fetchAll())},
		onCreateAccount: function(account) {dispatch(AccountActions.create(account))},
		onUpdateAccount: function(account) {dispatch(AccountActions.update(account))},
		onDeleteAccount: function(account) {dispatch(AccountActions.remove(account))},
		onSelectAccount: function(accountId) {dispatch(AccountActions.select(accountId))},
		onCreateTransaction: function(transaction) {dispatch(TransactionActions.create(transaction))},
		onUpdateTransaction: function(transaction) {dispatch(TransactionActions.update(transaction))},
		onDeleteTransaction: function(transaction) {dispatch(TransactionActions.remove(transaction))},
		onSelectTransaction: function(transactionId) {dispatch(TransactionActions.select(transactionId))},
		onUnselectTransaction: function() {dispatch(TransactionActions.unselect())},
		onFetchTransactionPage: function(account, pageSize, page) {dispatch(TransactionActions.fetchPage(account, pageSize, page))},
		onOpenImportModal: function() {dispatch(ImportActions.openModal())},
		onCloseImportModal: function() {dispatch(ImportActions.closeModal())},
		onImportOFX: function(account, password, startDate, endDate) {dispatch(ImportActions.importOFX(account, password, startDate, endDate))},
		onImportOFXFile: function(inputElement, account) {dispatch(ImportActions.importOFXFile(inputElement, account))},
		onImportGnucash: function(inputElement) {dispatch(ImportActions.importGnucash(inputElement))},
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(AccountsTab)
