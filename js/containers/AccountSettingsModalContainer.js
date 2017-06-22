var connect = require('react-redux').connect;

var UserActions = require('../actions/UserActions');

var AccountSettingsModal = require('../components/AccountSettingsModal');

function mapStateToProps(state) {
	return {
		user: state.user,
		currencies: state.securities.currency_list
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
