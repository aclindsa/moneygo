var assign = require('object-assign');

var TransactionConstants = require('../constants/TransactionConstants');
var UserConstants = require('../constants/UserConstants');

module.exports = function(state = {}, action) {
	switch (action.type) {
		case TransactionConstants.TRANSACTION_PAGE_FETCHED:
			var transactions = assign({}, state);
			for (var tidx in action.transactions) {
				var t = action.transactions[tidx];
				transactions = assign({}, transactions, {
					[t.TransactionId]: t
				});
			}
			return transactions;
		case TransactionConstants.TRANSACTION_CREATED:
		case TransactionConstants.TRANSACTION_UPDATED:
			var transaction = action.transaction;
			return assign({}, state, {
				[transaction.TransactionId]: transaction
			});
		case TransactionConstants.TRANSACTION_REMOVED:
			var transactions = assign({}, state);
			delete transactions[action.transactionId];
			return transactions;
		case UserConstants.USER_LOGGEDOUT:
			return {};
		default:
			return state;
	}
};
