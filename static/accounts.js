// Import all the objects we want to use from ReactBootstrap
var ListGroup = ReactBootstrap.ListGroup;
var ListGroupItem = ReactBootstrap.ListGroupItem;

var AccountList = React.createClass({
	getInitialState: function() {
		return {
		};
	},
	render: function() {
		var accounts = this.props.accounts;
		var account_map = this.props.account_map;

		var listGroupItems;

		for (var i = 0; i < accounts.length; i++) {
			listGroupItems += <ListGroupItem>{accounts[i].Name}</ListGroupItem>;
		}

		return (
			<ListGroup>
				{listGroupItems}
			</ListGroup>
		);
	}
});
