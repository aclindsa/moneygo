var assign = require('object-assign');

var SecurityTemplateConstants = require('../constants/SecurityTemplateConstants');
var UserConstants = require('../constants/UserConstants');

const initialState = {
	search: "",
	type: 0,
	templates: [],
	currencies: []
};

module.exports = function(state = initialState, action) {
	switch (action.type) {
		case SecurityTemplateConstants.SEARCH_SECURITY_TEMPLATES:
			return assign({}, state, {
				search: action.searchString,
				type: action.searchType,
				templates: []
			});
		case SecurityTemplateConstants.SECURITY_TEMPLATES_SEARCHED:
			if ((action.searchString != state.search) || (action.searchType != state.type))
				return state;
			return assign({}, state, {
				search: action.searchString,
				type: action.searchType,
				templates: action.securities
			});
		case SecurityTemplateConstants.CURRENCIES_FETCHED:
			return assign({}, state, {
				currencies: action.currencies
			});
		case UserConstants.USER_LOGGEDOUT:
			return assign({}, initialState, {
				currencies: state.currencies
			});
		default:
			return state;
	}
};
