var React = require('react');
var ReactDOM = require('react-dom');

var ReactBootstrap = require('react-bootstrap');
var Panel = ReactBootstrap.Panel;
var Form = ReactBootstrap.Form;
var FormGroup = ReactBootstrap.FormGroup;
var FormControl = ReactBootstrap.FormControl;
var ControlLabel = ReactBootstrap.ControlLabel;
var Col = ReactBootstrap.Col;
var Button = ReactBootstrap.Button;
var ButtonGroup = ReactBootstrap.ButtonGroup;

var models = require('./models.js');
var User = models.User;
var Error = models.Error;

module.exports = React.createClass({
	getInitialState: function() {
		return {error: "",
			name: "",
			username: "",
			email: "",
			password: "",
			confirm_password: "",
			passwordChanged: false,
			initial_password: ""};
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

		u.Name = this.state.name;
		u.Username = this.state.username;
		u.Email = this.state.email;
		u.Password = this.state.password;
		if (u.Password != this.state.confirm_password) {
			this.setState({error: "Error: password do not match"});
			return;
		}

		this.handleCreateNewUser(u);
	},
	handleCreateNewUser: function(user) {
		$.ajax({
			type: "POST",
			dataType: "json",
			url: "user/",
			data: {user: user.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					this.props.onNewUser();
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
		var title = <h3>Create New User</h3>;
		return (
			<Panel header={title} bsStyle="info">
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

					<ButtonGroup className="pull-right">
						<Button onClick={this.handleCancel}
								bsStyle="warning">Cancel</Button>
						<Button type="submit"
								bsStyle="success">Create New User</Button>
					</ButtonGroup>
				</Form>
			</Panel>
		);
	}
});
