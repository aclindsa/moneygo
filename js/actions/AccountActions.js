var AccountConstants = require('../constants/AccountConstants');

var ErrorActions = require('./ErrorActions');

var models = require('../models.js');
var Account = models.Account;
var Error = models.Error;

function fetchAccounts() {
	return {
		type: AccountConstants.FETCH_ACCOUNTS
	}
}

function accountsFetched(accounts) {
	return {
		type: AccountConstants.ACCOUNTS_FETCHED,
		accounts: accounts
	}
}

function createAccount() {
	return {
		type: AccountConstants.CREATE_ACCOUNT
	}
}

function accountCreated(account) {
	return {
		type: AccountConstants.ACCOUNT_CREATED,
		account: account
	}
}

function updateAccount() {
	return {
		type: AccountConstants.UPDATE_ACCOUNT
	}
}

function accountUpdated(account) {
	return {
		type: AccountConstants.ACCOUNT_UPDATED,
		account: account
	}
}

function removeAccount() {
	return {
		type: AccountConstants.REMOVE_ACCOUNT
	}
}

function accountRemoved(accountId) {
	return {
		type: AccountConstants.ACCOUNT_REMOVED,
		accountId: accountId
	}
}

function accountSelected(accountId) {
	return {
		type: AccountConstants.ACCOUNT_SELECTED,
		accountId: accountId
	}
}

function fetchAll() {
	return function (dispatch) {
		dispatch(fetchAccounts());

		$.ajax({
			type: "GET",
			dataType: "json",
			url: "account/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					dispatch(accountsFetched(data.accounts.map(function(json) {
						var a = new Account();
						a.fromJSON(json);
						return a;
					})));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function create(account) {
	return function (dispatch) {
		dispatch(createAccount());

		$.ajax({
			type: "POST",
			dataType: "json",
			url: "account/",
			data: {account: account.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					var a = new Account();
					a.fromJSON(data);
					dispatch(accountCreated(a));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function update(account) {
	return function (dispatch) {
		dispatch(updateAccount());

		$.ajax({
			type: "PUT",
			dataType: "json",
			url: "account/"+account.AccountId+"/",
			data: {account: account.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					var a = new Account();
					a.fromJSON(data);
					dispatch(accountUpdated(a));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function remove(account) {
	return function(dispatch) {
		dispatch(removeAccount());

		$.ajax({
			type: "DELETE",
			dataType: "json",
			url: "account/"+account.AccountId+"/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					dispatch(accountRemoved(account.AccountId));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

module.exports = {
	fetchAll: fetchAll,
	create: create,
	update: update,
	remove: remove,
	select: accountSelected
};
