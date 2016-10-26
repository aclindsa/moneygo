var assign = require('object-assign');

var SecurityTemplateConstants = require('../constants/SecurityTemplateConstants');
var UserConstants = require('../constants/UserConstants');

var SecurityType = require('../models').SecurityType;

module.exports = function(state = {search: "", type: 0, templates: [], searchNumber: 0}, action) {
	switch (action.type) {
		case SecurityTemplateConstants.SEARCH_SECURITY_TEMPLATES:
			return {
				search: action.searchString,
				type: action.searchType,
				templates: []
			};
		case SecurityTemplateConstants.SECURITY_TEMPLATES_SEARCHED:
			if ((action.searchString != state.search) || (action.searchType != state.type))
				return state;
			return {
				search: action.searchString,
				type: action.searchType,
				templates: action.securities
			};
		case UserConstants.USER_LOGGEDOUT:
			return {
				search: "",
				type: 0,
				templates: []
			};
		default:
			return state;
	}
};
