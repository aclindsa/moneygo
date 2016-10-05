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

function fetchAll() {
	return function (dispatch) {
		dispatch(fetchSecurities());

		$.ajax({
			type: "GET",
			dataType: "json",
			url: "security/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					ErrorActions.serverError(e);
				} else {
					dispatch(securitiesFetched(data.securities.map(function(json) {
						var s = new Security();
						s.fromJSON(json);
						return s;
					})));
				}
			},
			error: function(jqXHR, status, error) {
				ErrorActions.ajaxError(e);
			}
		});
	};
}

module.exports = {
	fetchAll: fetchAll
};
