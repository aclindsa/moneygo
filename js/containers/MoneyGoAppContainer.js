var connect = require('react-redux').connect;

var UserActions = require('../actions/UserActions');

var MoneyGoApp = require('../MoneyGoApp');

function mapStateToProps(state) {
	return {
		user: state.user
	}
}

function mapDispatchToProps(dispatch) {
	return {
		tryResumingSession: function() {dispatch(UserActions.tryResumingSession())},
	}
}

module.exports = connect(
	mapStateToProps,
	mapDispatchToProps
)(MoneyGoApp)
