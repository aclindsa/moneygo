var assign = require('object-assign');

var TransactionConstants = require('../constants/TransactionConstants');
var UserConstants = require('../constants/UserConstants');
var AccountConstants = require('../constants/AccountConstants');

var Account = require('../models').Account;

module.exports = function(state = {account: new Account(), pageSize: 1, page: 0, numPages: 0, transactions: [], endingBalance: "0", selection: -1, upToDate: false }, action) {
	switch (action.type) {
		case AccountConstants.ACCOUNT_SELECTED:
		case TransactionConstants.FETCH_TRANSACTION_PAGE:
			return assign({}, state, {
				account: action.account,
				pageSize: action.pageSize,
				page: action.page,
				numPages: 0,
				transactions: [],
				endingBalance: "0",
				upToDate: true
			});
		case TransactionConstants.TRANSACTION_PAGE_FETCHED:
			return assign({}, state, {
				account: action.account,
				pageSize: action.pageSize,
				page: action.page,
				numPages: action.numPages,
				transactions: action.transactions.map(function(t) {return t.TransactionId}),
				endingBalance: action.endingBalance,
				upToDate: true
			});
		case UserConstants.USER_LOGGEDOUT:
			return {
				account: new Account(),
				pageSize: 1,
				page: 0,
				numPages: 0,
				transactions: [],
				endingBalance: "0",
				selection: -1,
				upToDate: false
			};
		case TransactionConstants.TRANSACTION_CREATED:
		case TransactionConstants.TRANSACTION_UPDATED:
			return assign({}, state, {
				upToDate: false
			});
		case TransactionConstants.TRANSACTION_REMOVED:
			return assign({}, state, {
				transactions: state.transactions.filter(function(t) {return t != action.transactionId}),
				upToDate: false
			});
		case TransactionConstants.TRANSACTION_SELECTED:
			return assign({}, state, {
				selection: action.transactionId
			});
		case TransactionConstants.SELECTION_CLEARED:
			return assign({}, state, {
				selection: -1
			});
		default:
			return state;
	}
};
