var assign = require('object-assign');

var TransactionPageConstants = require('../constants/TransactionPageConstants');
var UserConstants = require('../constants/UserConstants');

var Account = require('../models').Account;

module.exports = function(state = {account: new Account(), pageSize: 1, page: 0, numPages: 0, transactions: [], endingBalance: "0" }, action) {
	switch (action.type) {
		case TransactionPageConstants.FETCH_TRANSACTION_PAGE:
			return {
				account: action.account,
				pageSize: action.pageSize,
				page: action.page,
				numPages: 0,
				transactions: [],
				endingBalance: "0"
			};
		case TransactionPageConstants.TRANSACTION_PAGE_FETCHED:
			return {
				account: action.account,
				pageSize: action.pageSize,
				page: action.page,
				numPages: action.numPages,
				transactions: action.transactions,
				endingBalance: action.endingBalance
			};
		case UserConstants.USER_LOGGEDOUT:
			return {
				account: new Account(),
				pageSize: 1,
				page: 0,
				numPages: 0,
				transactions: [],
				endingBalance: "0"
			};
		default:
			return state;
	}
};
