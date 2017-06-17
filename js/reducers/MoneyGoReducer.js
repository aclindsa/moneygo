var Redux = require('redux');

var UserReducer = require('./UserReducer');
var SessionReducer = require('./SessionReducer');
var AccountReducer = require('./AccountReducer');
var SecurityReducer = require('./SecurityReducer');
var SecurityTemplateReducer = require('./SecurityTemplateReducer');
var SelectedAccountReducer = require('./SelectedAccountReducer');
var SelectedSecurityReducer = require('./SelectedSecurityReducer');
var ReportReducer = require('./ReportReducer');
var TransactionReducer = require('./TransactionReducer');
var TransactionPageReducer = require('./TransactionPageReducer');
var ImportReducer = require('./ImportReducer');
var ErrorReducer = require('./ErrorReducer');

module.exports = Redux.combineReducers({
	user: UserReducer,
	session: SessionReducer,
	accounts: AccountReducer,
	securities: SecurityReducer,
	securityTemplates: SecurityTemplateReducer,
	selectedAccount: SelectedAccountReducer,
	selectedSecurity: SelectedSecurityReducer,
	reports: ReportReducer,
	transactions: TransactionReducer,
	transactionPage: TransactionPageReducer,
	imports: ImportReducer,
	error: ErrorReducer
});
