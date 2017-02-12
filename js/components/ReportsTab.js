var React = require('react');
var ReactDOM = require('react-dom');

var ReactBootstrap = require('react-bootstrap');
var Grid = ReactBootstrap.Grid;
var Row = ReactBootstrap.Row;
var Col = ReactBootstrap.Col;
var Form = ReactBootstrap.Form;
var FormGroup = ReactBootstrap.FormGroup;
var FormControl = ReactBootstrap.FormControl;
var ControlLabel = ReactBootstrap.ControlLabel;
var Button = ReactBootstrap.Button;
var ButtonGroup = ReactBootstrap.ButtonGroup;
var ButtonToolbar = ReactBootstrap.ButtonToolbar;
var Glyphicon = ReactBootstrap.Glyphicon;
var ListGroup = ReactBootstrap.ListGroup;
var ListGroupItem = ReactBootstrap.ListGroupItem;
var Modal = ReactBootstrap.Modal;
var Panel = ReactBootstrap.Panel;

var Combobox = require('react-widgets').Combobox;

module.exports = React.createClass({
	displayName: "ReportsTab",
	getInitialState: function() {
		return { };
	},
	componentWillMount: function() {
		this.props.onFetchReport("monthly_expenses");
	},
	render: function() {
		console.log(this.props.reports);
		return (
			<div>hello</div>
		);
	}
});
