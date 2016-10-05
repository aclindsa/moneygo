var connect = require('react-redux').connect;

var UserActions = require('../actions/UserActions');
var ErrorActions = require('../actions/ErrorActions');

var TopBar = require('../components/TopBar');

function mapStateToProps(state) {
	return {
		user: state.user,
		error: state.error
	}
}

function mapDispatchToProps(dispatch) {
	return {
		onLogin: function(user) {dispatch(UserActions.login(user))},
		onLogout: function() {dispatch(UserActions.logout())},
		onUpdateUser: function(user) {dispatch(UserActions.update(user))},
		onClearError: function() {dispatch(ErrorActions.clearError())}
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(TopBar)
