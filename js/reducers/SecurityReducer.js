var assign = require('object-assign');

var SecurityConstants = require('../constants/SecurityConstants');
var UserConstants = require('../constants/UserConstants');

var SecurityType = require('../models').SecurityType;

const initialState = {
	map: {},
	list: [],
	currency_list: []
}

function mapToList(securities) {
	var security_list = [];
	for (var securityId in securities) {
		if (securities.hasOwnProperty(securityId))
			security_list.push(securities[securityId]);
	}
	return security_list;
}

function mapToCurrencyList(securities) {
	var security_list = [];
	for (var securityId in securities) {
		if (securities.hasOwnProperty(securityId) && securities[securityId].Type == SecurityType.Currency)
			security_list.push(securities[securityId]);
	}
	return security_list;
}

module.exports = function(state = initialState, action) {
	switch (action.type) {
		case SecurityConstants.SECURITIES_FETCHED:
			var securities = {};
			var list = [];
			var currency_list = [];
			for (var i = 0; i < action.securities.length; i++) {
				var security = action.securities[i];
				securities[security.SecurityId] = security;
				list.push(security);
				if (security.Type == SecurityType.Currency)
					currency_list.push(security);
			}
			return {
				map: securities,
				list: list,
				currency_list: currency_list
			};
		case SecurityConstants.SECURITY_CREATED:
		case SecurityConstants.SECURITY_UPDATED:
			var security = action.security;
			var map = assign({}, state.map, {
				[security.SecurityId]: security
			});
			return {
				map: map,
				list: mapToList(map),
				currency_list: mapToCurrencyList(map)
			};
		case SecurityConstants.SECURITY_REMOVED:
			var map = assign({}, state.map);
			delete map[action.securityId];
			return {
				map: map,
				list: mapToList(map),
				currency_list: mapToCurrencyList(map)
			};
		case UserConstants.USER_LOGGEDOUT:
			return initialState;
		default:
			return state;
	}
};
