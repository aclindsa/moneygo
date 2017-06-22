var connect = require('react-redux').connect;

var UserActions = require('../actions/UserActions');
var SecurityTemplateActions = require('../actions/SecurityTemplateActions');

var MoneyGoApp = require('../components/MoneyGoApp');

function mapStateToProps(state) {
	return {
		user: state.user
	}
}

function mapDispatchToProps(dispatch) {
	return {
		tryResumingSession: function() {dispatch(UserActions.tryResumingSession())},
		fetchCurrencies: function() {dispatch(SecurityTemplateActions.fetchCurrencies())},
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(MoneyGoApp)
