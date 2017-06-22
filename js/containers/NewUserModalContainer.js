var connect = require('react-redux').connect;

var UserActions = require('../actions/UserActions');

var NewUserModal = require('../components/NewUserModal');

function mapStateToProps(state) {
	return {
		currencies: state.securityTemplates.currencies
	}
}

function mapDispatchToProps(dispatch) {
	return {
		createNewUser: function(user) {dispatch(UserActions.create(user))}
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(NewUserModal)
