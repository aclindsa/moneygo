var connect = require('react-redux').connect;

var UserActions = require('../actions/UserActions');
var AccountSettingsModal = require('../components/AccountSettingsModal');

function mapStateToProps(state) {
	return {
		user: state.user
	}
}

function mapDispatchToProps(dispatch) {
	return {
		onUpdateUser: function(user) {dispatch(UserActions.update(user))}
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(AccountSettingsModal)
