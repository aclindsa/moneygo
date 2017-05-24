var TransactionConstants = require('../constants/TransactionConstants');

var ErrorActions = require('./ErrorActions');

var models = require('../models.js');
var Account = models.Account;
var Transaction = models.Transaction;
var Error = models.Error;

var Big = require('big.js');

function fetchTransactionPage(account, pageSize, page) {
	return {
		type: TransactionConstants.FETCH_TRANSACTION_PAGE,
		account: account,
		pageSize: pageSize,
		page: page
	}
}

function transactionPageFetched(account, pageSize, page, numPages,
		transactions, endingBalance) {
	return {
		type: TransactionConstants.TRANSACTION_PAGE_FETCHED,
		account: account,
		pageSize: pageSize,
		page: page,
		numPages: numPages,
		transactions: transactions,
		endingBalance: endingBalance
	}
}

function createTransaction() {
	return {
		type: TransactionConstants.CREATE_TRANSACTION
	}
}

function transactionCreated(transaction) {
	return {
		type: TransactionConstants.TRANSACTION_CREATED,
		transaction: transaction
	}
}

function updateTransaction() {
	return {
		type: TransactionConstants.UPDATE_TRANSACTION
	}
}

function transactionUpdated(transaction) {
	return {
		type: TransactionConstants.TRANSACTION_UPDATED,
		transaction: transaction
	}
}

function removeTransaction() {
	return {
		type: TransactionConstants.REMOVE_TRANSACTION
	}
}

function transactionRemoved(transactionId) {
	return {
		type: TransactionConstants.TRANSACTION_REMOVED,
		transactionId: transactionId
	}
}

function transactionSelected(transactionId) {
	return {
		type: TransactionConstants.TRANSACTION_SELECTED,
		transactionId: transactionId
	}
}

function selectionCleared() {
	return {
		type: TransactionConstants.SELECTION_CLEARED
	}
}

function fetchPage(account, pageSize, page) {
	return function (dispatch) {
		dispatch(fetchTransactionPage(account, pageSize, page));

		$.ajax({
			type: "GET",
			dataType: "json",
			url: "account/"+account.AccountId+"/transactions?sort=date-desc&limit="+pageSize+"&page="+page,
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
					return;
				}

				var transactions = [];
				var balance = new Big(data.EndingBalance);

				for (var i = 0; i < data.Transactions.length; i++) {
					var t = new Transaction();
					t.fromJSON(data.Transactions[i]);

					t.Balance = balance.plus(0); // Make a copy of the current balance
					// Keep a talley of the running balance of these transactions
					for (var j = 0; j < data.Transactions[i].Splits.length; j++) {
						var split = data.Transactions[i].Splits[j];
						if (account.AccountId == split.AccountId) {
							balance = balance.minus(split.Amount);
						}
					}
					transactions.push(t);
				}
				var a = new Account();
				a.fromJSON(data.Account);

				var numPages = Math.ceil(data.TotalTransactions / pageSize);

				dispatch(transactionPageFetched(account, pageSize, page,
					numPages, transactions, new Big(data.EndingBalance)));
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function create(transaction) {
	return function (dispatch) {
		dispatch(createTransaction());

		$.ajax({
			type: "POST",
			dataType: "json",
			url: "transaction/",
			data: {transaction: transaction.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					var t = new Transaction();
					t.fromJSON(data);
					dispatch(transactionCreated(t));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function update(transaction) {
	return function (dispatch) {
		dispatch(updateTransaction());

		$.ajax({
			type: "PUT",
			dataType: "json",
			url: "transaction/"+transaction.TransactionId+"/",
			data: {transaction: transaction.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					var t = new Transaction();
					t.fromJSON(data);
					dispatch(transactionUpdated(t));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function remove(transaction) {
	return function(dispatch) {
		dispatch(removeTransaction());

		$.ajax({
			type: "DELETE",
			dataType: "json",
			url: "transaction/"+transaction.TransactionId+"/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					dispatch(transactionRemoved(transaction.TransactionId));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

module.exports = {
	fetchPage: fetchPage,
	create: create,
	update: update,
	remove: remove,
	select: transactionSelected,
	unselect: selectionCleared
};
