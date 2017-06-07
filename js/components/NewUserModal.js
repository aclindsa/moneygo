var React = require('react');
var ReactDOM = require('react-dom');

var ReactBootstrap = require('react-bootstrap');
var Modal = ReactBootstrap.Modal;
var Form = ReactBootstrap.Form;
var FormGroup = ReactBootstrap.FormGroup;
var FormControl = ReactBootstrap.FormControl;
var ControlLabel = ReactBootstrap.ControlLabel;
var Col = ReactBootstrap.Col;
var Button = ReactBootstrap.Button;
var ButtonGroup = ReactBootstrap.ButtonGroup;

var models = require('../models');
var User = models.User;

class NewUserModal extends React.Component {
	constructor() {
		super();
		this.state = {
			error: "",
			name: "",
			username: "",
			email: "",
			password: "",
			confirm_password: "",
			passwordChanged: false,
			initial_password: ""
		};
	}
	passwordValidationState() {
		if (this.state.passwordChanged) {
			if (this.state.password.length >= 10)
				return "success";
			else if (this.state.password.length >= 6)
				return "warning";
			else
				return "error";
		}
	}
	confirmPasswordValidationState() {
		if (this.state.confirm_password.length > 0) {
			if (this.state.confirm_password == this.state.password)
				return "success";
			else
				return "error";
		}
	}
	handleCancel() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	}
	handleChange() {
		if (ReactDOM.findDOMNode(this.refs.password).value != this.state.initial_password)
			this.setState({passwordChanged: true});
		this.setState({
			name: ReactDOM.findDOMNode(this.refs.name).value,
			username: ReactDOM.findDOMNode(this.refs.username).value,
			email: ReactDOM.findDOMNode(this.refs.email).value,
			password: ReactDOM.findDOMNode(this.refs.password).value,
			confirm_password: ReactDOM.findDOMNode(this.refs.confirm_password).value
		});
	}
	handleSubmit(e) {
		var u = new User();
		var error = "";
		e.preventDefault();

		u.Name = this.state.name;
		u.Username = this.state.username;
		u.Email = this.state.email;
		u.Password = this.state.password;
		if (u.Password != this.state.confirm_password) {
			this.setState({error: "Error: passwords do not match"});
			return;
		}

		this.props.createNewUser(u);
		if (this.props.onSubmit != null)
			this.props.onSubmit(u);
	}
	render() {
		return (
			<Modal show={this.props.show} onHide={this.handleCancel} bsSize="large">
				<Modal.Header closeButton>
					<Modal.Title>Create New user</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<span style={{color: "red"}}>{this.state.error}</span>
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
						<Button onClick={this.handleCancel}
								bsStyle="warning">Cancel</Button>
						<Button onClick={this.handleSubmit}
								bsStyle="success">Create New User</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
}

module.exports = NewUserModal;
