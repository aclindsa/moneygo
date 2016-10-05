var React = require('react');

var ReactBootstrap = require('react-bootstrap');
var Alert = ReactBootstrap.Alert;
var FormGroup = ReactBootstrap.FormGroup;
var FormControl = ReactBootstrap.FormControl;
var Button = ReactBootstrap.Button;
var DropdownButton = ReactBootstrap.DropdownButton;
var MenuItem = ReactBootstrap.MenuItem;
var Row = ReactBootstrap.Row;
var Col = ReactBootstrap.Col;

var ReactDOM = require('react-dom');

var User = require('./models.js').User;

const LoginBar = React.createClass({
	getInitialState: function() {
		return {username: '', password: ''};
	},
	onUsernameChange: function(e) {
		this.setState({username: e.target.value});
	},
	onPasswordChange: function(e) {
		this.setState({password: e.target.value});
	},
	handleSubmit: function(e) {
		var user = new User();
		e.preventDefault();
		user.Username = ReactDOM.findDOMNode(this.refs.username).value;
		user.Password = ReactDOM.findDOMNode(this.refs.password).value;
		this.props.onLogin(user);
	},
	handleNewUserSubmit: function(e) {
		e.preventDefault();
		this.props.onCreateNewUser();
	},
	render: function() {
		return (
			<form onSubmit={this.handleSubmit}>
			<FormGroup>
				<Row>
					<Col xs={4}></Col>
					<Col xs={2}>
						<Button bsStyle="link"
							onClick={this.handleNewUserSubmit}>Create New User</Button>
					</Col>
					<Col xs={2}>
						<FormControl type="text"
							placeholder="Username..."
							ref="username"/>
					</Col>
					<Col xs={2}>
						<FormControl type="password"
							placeholder="Password..."
							ref="password"/>
					</Col>
					<Col xs={2}>
						<Button type="submit" bsStyle="primary" block>
							Login</Button>
					</Col>
				</Row>
			</FormGroup>
			</form>
		);
	}
});

const LogoutBar = React.createClass({
	handleOnSelect: function(key) {
		if (key == 1) {
			if (this.props.onAccountSettings != null)
				this.props.onAccountSettings();
		} else if (key == 2) {
			this.props.onLogout();
		}
	},
	render: function() {
		var signedInString = "Signed in as "+this.props.user.Name;
		return (
			<FormGroup>
				<Row>
					<Col xs={2}><label className="control-label pull-left">Money<i>Go</i></label></Col>
					<Col xs={6}></Col>
					<Col xs={4}>
						<div className="pull-right">
						<DropdownButton id="logout-settings-dropdown" title={signedInString} onSelect={this.handleOnSelect} bsStyle="info">
							<MenuItem eventKey={1}>Account Settings</MenuItem>
							<MenuItem eventKey={2}>Logout</MenuItem>
						</DropdownButton>
						</div>
					</Col>
				</Row>
			</FormGroup>
		);
	}
});

module.exports = React.createClass({
	displayName: "TopBar",
	render: function() {
		var barContents;
		var errorAlert;
		if (!this.props.user.isUser())
			barContents = <LoginBar onLogin={this.props.onLogin} onCreateNewUser={this.props.onCreateNewUser} />;
		else
			barContents = <LogoutBar user={this.props.user} onLogout={this.props.onLogout} onAccountSettings={this.props.onAccountSettings}/>;
		if (this.props.error.isError())
			errorAlert =
					<Alert bsStyle="danger" onDismiss={this.props.onClearError}>
						<h4>Error!</h4>
						<p>Error {this.props.error.ErrorId}: {this.props.error.ErrorString}</p>
						<Button onClick={this.props.onClearError}>Clear</Button>
					</Alert>;

		return (
			<div>
				{barContents}
				{errorAlert}
			</div>
		);
	}
});
