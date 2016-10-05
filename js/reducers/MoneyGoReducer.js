var Redux = require('redux');

var UserReducer = require('./UserReducer');
var SessionReducer = require('./SessionReducer');
var AccountReducer = require('./AccountReducer');
var SecurityReducer = require('./SecurityReducer');
var SelectedAccountReducer = require('./SelectedAccountReducer');
var ErrorReducer = require('./ErrorReducer');

module.exports = Redux.combineReducers({
	user: UserReducer,
	session: SessionReducer,
	accounts: AccountReducer,
	securities: SecurityReducer,
	selectedAccount: SelectedAccountReducer,
	error: ErrorReducer
});
