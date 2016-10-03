var React = require('react');
var ReactDOM = require('react-dom');

var Provider = require('react-redux').Provider;
var Redux = require('redux');
var ReduxThunk = require('redux-thunk').default;

var Globalize = require('globalize');
var globalizeLocalizer = require('react-widgets/lib/localizers/globalize');

var MoneyGoApp = require('./MoneyGoApp.js');
var MoneyGoReducer = require('./reducers/MoneyGoReducer');

// Setup globalization for react-widgets
//Globalize.load(require("cldr-data").entireSupplemental());
Globalize.load(
	require("cldr-data/main/en/ca-gregorian"),
	require("cldr-data/main/en/numbers"),
	require("cldr-data/supplemental/likelySubtags"),
	require("cldr-data/supplemental/timeData"),
	require("cldr-data/supplemental/weekData")
);
Globalize.locale('en');
globalizeLocalizer(Globalize);

$(document).ready(function() {
	var store = Redux.createStore(
		MoneyGoReducer,
		Redux.applyMiddleware(
			ReduxThunk
		)
	);

	ReactDOM.render(
		<Provider store={store}>
			<MoneyGoApp />
		</Provider>,
		document.getElementById("content")
	);
});
