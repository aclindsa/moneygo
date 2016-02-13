var React = require('react');

var Modal = require('react-bootstrap').Modal;
var Button = require('react-bootstrap').Button;
var ButtonGroup = require('react-bootstrap').ButtonGroup;
var Input = require('react-bootstrap').Input;

module.exports = React.createClass({
	displayName: "AccountSettingsModal",
	_getInitialState: function(props) {
		return {error: "",
			name: props.user.Name,
			username: props.user.Username,
			email: props.user.Email,
			password: BogusPassword,
			confirm_password: BogusPassword,
			passwordChanged: false,
			initial_password: BogusPassword};
	},
	getInitialState: function() {
		 return this._getInitialState(this.props);
	},
	componentWillReceiveProps: function(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this._getInitialState(nextProps));
		}
	},
	passwordValidationState: function() {
		if (this.state.passwordChanged) {
			if (this.state.password.length >= 10)
				return "success";
			else if (this.state.password.length >= 6)
				return "warning";
			else
				return "error";
		}
	},
	confirmPasswordValidationState: function() {
		if (this.state.confirm_password.length > 0) {
			if (this.state.confirm_password == this.state.password)
				return "success";
			else
				return "error";
		}
	},
	handleCancel: function() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	},
	handleChange: function() {
		if (this.refs.password.getValue() != this.state.initial_password)
			this.setState({passwordChanged: true});
		this.setState({
			name: this.refs.name.getValue(),
			username: this.refs.username.getValue(),
			email: this.refs.email.getValue(),
			password: this.refs.password.getValue(),
			confirm_password: this.refs.confirm_password.getValue()
		});
	},
	handleSubmit: function(e) {
		var u = new User();
		var error = "";
		e.preventDefault();

		u.UserId = this.props.user.UserId;
		u.Name = this.state.name;
		u.Username = this.state.username;
		u.Email = this.state.email;
		if (this.state.passwordChanged) {
			u.Password = this.state.password;
			if (u.Password != this.state.confirm_password) {
				this.setState({error: "Error: password do not match"});
				return;
			}
		} else {
			u.Password = BogusPassword;
		}

		this.handleSaveSettings(u);
	},
	handleSaveSettings: function(user) {
		$.ajax({
			type: "PUT",
			dataType: "json",
			url: "user/"+user.UserId+"/",
			data: {user: user.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					user.Password = "";
					this.props.onSubmit(user);
				}
			}.bind(this),
			error: function(jqXHR, status, error) {
				var e = new Error();
				e.ErrorId = 5;
				e.ErrorString = "Request Failed: " + status + error;
				this.setState({error: e});
			}.bind(this),
		});
	},
	render: function() {
		return (
			<Modal show={this.props.show} onHide={this.handleCancel} bsSize="large">
				<Modal.Header closeButton>
					<Modal.Title>Edit Account Settings</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<span color="red">{this.state.error}</span>
				<form onSubmit={this.handleSubmit}
						className="form-horizontal">
					<Input type="text"
							label="Name"
							value={this.state.name}
							onChange={this.handleChange}
							ref="name"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"/>
					<Input type="text"
							label="Username"
							value={this.state.username}
							onChange={this.handleChange}
							ref="username"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"/>
					<Input type="email"
							label="Email"
							value={this.state.email}
							onChange={this.handleChange}
							ref="email"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"/>
					<Input type="password"
							label="Password"
							value={this.state.password}
							onChange={this.handleChange}
							ref="password"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"
							bsStyle={this.passwordValidationState()}
							hasFeedback/>
					<Input type="password"
							label="Confirm Password"
							value={this.state.confirm_password}
							onChange={this.handleChange}
							ref="confirm_password"
							labelClassName="col-xs-2"
							wrapperClassName="col-xs-10"
							bsStyle={this.confirmPasswordValidationState()}
							hasFeedback/>
				</form>
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup>
						<Button onClick={this.handleCancel} bsStyle="warning">Cancel</Button>
						<Button onClick={this.handleSubmit} bsStyle="success">Save Settings</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
});
