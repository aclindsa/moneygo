var SecurityTemplateConstants = require('../constants/SecurityTemplateConstants');

var ErrorActions = require('./ErrorActions');

var models = require('../models.js');
var Security = models.Security;
var Error = models.Error;
var SecurityType = models.SecurityType;

function searchSecurityTemplates(searchString, searchType) {
	return {
		type: SecurityTemplateConstants.SEARCH_SECURITY_TEMPLATES,
		searchString: searchString,
		searchType: searchType
	}
}

function securityTemplatesSearched(searchString, searchType, securities) {
	return {
		type: SecurityTemplateConstants.SECURITY_TEMPLATES_SEARCHED,
		searchString: searchString,
		searchType: searchType,
		securities: securities
	}
}

function fetchCurrencyTemplates() {
	return {
		type: SecurityTemplateConstants.FETCH_CURRENCIES
	}
}

function currencyTemplatesFetched(currencies) {
	return {
		type: SecurityTemplateConstants.CURRENCIES_FETCHED,
		currencies: currencies
	}
}

function search(searchString, searchType, limit) {
	return function (dispatch) {
		dispatch(searchSecurityTemplates(searchString, searchType));

		if (searchString == "")
			return;

		$.ajax({
			type: "GET",
			dataType: "json",
			url: "securitytemplate/?search="+searchString+"&type="+searchType+"&limit="+limit,
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else if (data.securities == null) {
					dispatch(securityTemplatesSearched(searchString, searchType, new Array()));
				} else {
					dispatch(securityTemplatesSearched(searchString, searchType,
							data.securities.map(function(json) {
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

function fetchCurrencies() {
	return function (dispatch) {
		dispatch(fetchCurrencyTemplates());

		$.ajax({
			type: "GET",
			dataType: "json",
			url: "securitytemplate/?search=&type=currency",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					dispatch(ErrorActions.serverError(e));
				} else if (data.securities == null) {
					dispatch(currencyTemplatesFetched(new Array()));
				} else {
					dispatch(currencyTemplatesFetched(
							data.securities.map(function(json) {
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

module.exports = {
	search: search,
	fetchCurrencies: fetchCurrencies
};
