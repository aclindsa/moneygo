var React = require('react');

var ReactDOM = require('react-dom');

var ReactBootstrap = require('react-bootstrap');
var Modal = ReactBootstrap.Modal;
var Button = ReactBootstrap.Button;
var ButtonGroup = ReactBootstrap.ButtonGroup;
var Form = ReactBootstrap.Form;
var FormGroup = ReactBootstrap.FormGroup;
var FormControl = ReactBootstrap.FormControl;
var ControlLabel = ReactBootstrap.ControlLabel;
var Col = ReactBootstrap.Col;

var models = require('./models.js');
var User = models.User;
var Error = models.Error;

module.exports = React.createClass({
	displayName: "AccountSettingsModal",
	_getInitialState: function(props) {
		return {error: "",
			name: props.user.Name,
			username: props.user.Username,
			email: props.user.Email,
			password: models.BogusPassword,
			confirm_password: models.BogusPassword,
			passwordChanged: false,
			initial_password: models.BogusPassword};
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
		if (ReactDOM.findDOMNode(this.refs.password).value != this.state.initial_password)
			this.setState({passwordChanged: true});
		this.setState({
			name: ReactDOM.findDOMNode(this.refs.name).value,
			username: ReactDOM.findDOMNode(this.refs.username).value,
			email: ReactDOM.findDOMNode(this.refs.email).value,
			password: ReactDOM.findDOMNode(this.refs.password).value,
			confirm_password: ReactDOM.findDOMNode(this.refs.confirm_password).value
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
			u.Password = models.BogusPassword;
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
				<Form horizontal onSubmit={this.handleSubmit}>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Name</Col>
						<Col xs={10}>
						<FormControl type="text"
								value={this.state.name}
								onChange={this.handleChange}
								ref="name"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Username</Col>
						<Col xs={10}>
						<FormControl type="text"
							value={this.state.username}
							onChange={this.handleChange}
							ref="username"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Email</Col>
						<Col xs={10}>
						<FormControl type="email"
							value={this.state.email}
							onChange={this.handleChange}
							ref="email"/>
						</Col>
					</FormGroup>
					<FormGroup validationState={this.passwordValidationState()}>
						<Col componentClass={ControlLabel} xs={2}>Password</Col>
						<Col xs={10}>
						<FormControl type="password"
							value={this.state.password}
							onChange={this.handleChange}
							ref="password"/>
						<FormControl.Feedback/>
						</Col>
					</FormGroup>
					<FormGroup validationState={this.confirmPasswordValidationState()}>
						<Col componentClass={ControlLabel} xs={2}>Confirm Password</Col>
						<Col xs={10}>
						<FormControl type="password"
							value={this.state.confirm_password}
							onChange={this.handleChange}
							ref="confirm_password"/>
						<FormControl.Feedback/>
						</Col>
					</FormGroup>
				</Form>
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
