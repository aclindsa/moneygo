var Redux = require('redux');

var AccountReducer = require('./AccountReducer');
var SelectedAccountReducer = require('./SelectedAccountReducer');

module.exports = Redux.combineReducers({
	accounts: AccountReducer,
	selectedAccount: SelectedAccountReducer
});
