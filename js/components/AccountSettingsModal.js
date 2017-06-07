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

var User = require('../models').User;

class AccountSettingsModal extends React.Component {
	_getInitialState(props) {
		return {
			error: "",
			name: props ? props.user.Name: "",
			username: props ? props.user.Username : "",
			email: props ? props.user.Email : "",
			password: models.BogusPassword,
			confirm_password: models.BogusPassword,
			passwordChanged: false,
			initial_password: models.BogusPassword
		};
	}
	constructor() {
		super();
		this.state = this._getInitialState();
		this.onCancel = this.handleCancel.bind(this);
		this.onChange = this.handleChange.bind(this);
		this.onSubmit = this.handleSubmit.bind(this);
	}
	componentWillReceiveProps(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this._getInitialState(nextProps));
		}
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

		this.props.onUpdateUser(u);
		this.props.onSubmit();
	}
	render() {
		return (
			<Modal show={this.props.show} onHide={this.onCancel} bsSize="large">
				<Modal.Header closeButton>
					<Modal.Title>Edit Account Settings</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<span color="red">{this.state.error}</span>
				<Form horizontal onSubmit={this.onSubmit}>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Name</Col>
						<Col xs={10}>
						<FormControl type="text"
								value={this.state.name}
								onChange={this.onChange}
								ref="name"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Username</Col>
						<Col xs={10}>
						<FormControl type="text"
							value={this.state.username}
							onChange={this.onChange}
							ref="username"/>
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Email</Col>
						<Col xs={10}>
						<FormControl type="email"
							value={this.state.email}
							onChange={this.onChange}
							ref="email"/>
						</Col>
					</FormGroup>
					<FormGroup validationState={this.passwordValidationState()}>
						<Col componentClass={ControlLabel} xs={2}>Password</Col>
						<Col xs={10}>
						<FormControl type="password"
							value={this.state.password}
							onChange={this.onChange}
							ref="password"/>
						<FormControl.Feedback/>
						</Col>
					</FormGroup>
					<FormGroup validationState={this.confirmPasswordValidationState()}>
						<Col componentClass={ControlLabel} xs={2}>Confirm Password</Col>
						<Col xs={10}>
						<FormControl type="password"
							value={this.state.confirm_password}
							onChange={this.onChange}
							ref="confirm_password"/>
						<FormControl.Feedback/>
						</Col>
					</FormGroup>
				</Form>
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup>
						<Button onClick={this.onCancel} bsStyle="warning">Cancel</Button>
						<Button onClick={this.onSubmit} bsStyle="success">Save Settings</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
}

module.exports = AccountSettingsModal;
