var SecurityConstants = require('../constants/SecurityConstants');

var ErrorActions = require('./ErrorActions');

var models = require('../models.js');
var Security = models.Security;
var Error = models.Error;

function fetchSecurities() {
	return {
		type: SecurityConstants.FETCH_SECURITIES
	}
}

function securitiesFetched(securities) {
	return {
		type: SecurityConstants.SECURITIES_FETCHED,
		securities: securities
	}
}

function createSecurity() {
	return {
		type: SecurityConstants.CREATE_SECURITY
	}
}

function securityCreated(security) {
	return {
		type: SecurityConstants.SECURITY_CREATED,
		security: security
	}
}

function updateSecurity() {
	return {
		type: SecurityConstants.UPDATE_SECURITY
	}
}

function securityUpdated(security) {
	return {
		type: SecurityConstants.SECURITY_UPDATED,
		security: security
	}
}

function removeSecurity() {
	return {
		type: SecurityConstants.REMOVE_SECURITY
	}
}

function securityRemoved(securityId) {
	return {
		type: SecurityConstants.SECURITY_REMOVED,
		securityId: securityId
	}
}

function securitySelected(securityId) {
	return {
		type: SecurityConstants.SECURITY_SELECTED,
		securityId: securityId
	}
}

function fetchAll() {
	return function (dispatch) {
		dispatch(fetchSecurities());

		$.ajax({
			type: "GET",
			dataType: "json",
			url: "v1/securities/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					dispatch(securitiesFetched(data.securities.map(function(json) {
						var s = new Security();
						s.fromJSON(json);
						return s;
					})));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function create(security) {
	return function (dispatch) {
		dispatch(createSecurity());

		$.ajax({
			type: "POST",
			dataType: "json",
			url: "v1/securities/",
			data: {security: security.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					var s = new Security();
					s.fromJSON(data);
					dispatch(securityCreated(s));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function update(security) {
	return function (dispatch) {
		dispatch(updateSecurity());

		$.ajax({
			type: "PUT",
			dataType: "json",
			url: "v1/securities/"+security.SecurityId+"/",
			data: {security: security.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					var s = new Security();
					s.fromJSON(data);
					dispatch(securityUpdated(s));
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(ErrorActions.ajaxError(error));
			}
		});
	};
}

function remove(security) {
	return function(dispatch) {
		dispatch(removeSecurity());

		$.ajax({
			type: "DELETE",
			dataType: "json",
			url: "v1/securities/"+security.SecurityId+"/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else {
					dispatch(securityRemoved(security.SecurityId));
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
	select: securitySelected
};
