var connect = require('react-redux').connect;

var SecurityActions = require('../actions/SecurityActions');
var SecurityTemplateActions = require('../actions/SecurityTemplateActions');
var SecuritiesTab = require('../components/SecuritiesTab');

function mapStateToProps(state) {
	var selectedSecurityAccounts = [];
	for (var accountId in state.accounts.map) {
		if (state.accounts.map.hasOwnProperty(accountId)
				&& state.accounts.map[accountId].SecurityId == state.selectedSecurity)
			selectedSecurityAccounts.push(state.accounts.map[accountId]);
	}
	return {
		securities: state.securities.map,
		security_list: state.securities.list,
		selectedSecurityAccounts: selectedSecurityAccounts,
		selectedSecurity: state.selectedSecurity,
		securityTemplates: state.securityTemplates
	}
}

function mapDispatchToProps(dispatch) {
	return {
		onCreateSecurity: function(security) {dispatch(SecurityActions.create(security))},
		onUpdateSecurity: function(security) {dispatch(SecurityActions.update(security))},
		onDeleteSecurity: function(securityId) {dispatch(SecurityActions.remove(securityId))},
		onSelectSecurity: function(securityId) {dispatch(SecurityActions.select(securityId))},
		onSearchTemplates: function(search, type, limit) {dispatch(SecurityTemplateActions.search(search, type, limit))}
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(SecuritiesTab)
