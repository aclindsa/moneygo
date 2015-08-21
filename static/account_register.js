// Import all the objects we want to use from ReactBootstrap

var Modal = ReactBootstrap.Modal;
var Pagination = ReactBootstrap.Pagination;

var Label = ReactBootstrap.Label;
var Table = ReactBootstrap.Table;
var Grid = ReactBootstrap.Grid;
var Row = ReactBootstrap.Row;
var Col = ReactBootstrap.Col;

var Button = ReactBootstrap.Button;
var ButtonToolbar = ReactBootstrap.ButtonToolbar;

var DateTimePicker = ReactWidgets.DateTimePicker;

const TransactionRow = React.createClass({
	handleClick: function(e) {
		const refs = ["date", "number", "description", "account", "status", "amount"];
		for (var ref in refs) {
			if (this.refs[refs[ref]].getDOMNode() == e.target) {
				this.props.onEdit(this.props.transaction, refs[ref]);
				return;
			}
		}
	},
	render: function() {
		var date = this.props.transaction.Date;
		var dateString = date.getFullYear() + "/" + (date.getMonth()+1) + "/" + date.getDate();
		var number = ""
		var accountName = "";
		var status = "";
		var security = this.props.security_map[this.props.account.SecurityId];

		if (this.props.transaction.isTransaction()) {
			var thisAccountSplit;
			for (var i = 0; i < this.props.transaction.Splits.length; i++) {
				if (this.props.transaction.Splits[i].AccountId == this.props.account.AccountId) {
					thisAccountSplit = this.props.transaction.Splits[i];
					break;
				}
			}
			if (this.props.transaction.Splits.length == 2) {
				var otherSplit = this.props.transaction.Splits[0];
				if (otherSplit.AccountId == this.props.account.AccountId)
					var otherSplit = this.props.transaction.Splits[1];
				var accountName = getAccountDisplayName(this.props.account_map[otherSplit.AccountId], this.props.account_map);
			} else {
				accountName = "--Split Transaction--";
			}

			var amount = "$" + thisAccountSplit.Amount.toFixed(security.Precision);
			var balance = "$" + this.props.transaction.Balance.toFixed(security.Precision);
			status = TransactionStatusMap[this.props.transaction.Status];
			number = thisAccountSplit.Number;
		} else {
			var amount = "$" + (new Big(0.0)).toFixed(security.Precision);
			var balance = "$" + (new Big(0.0)).toFixed(security.Precision);
		}

		return (
			<tr>
				<td ref="date" onClick={this.handleClick}>{dateString}</td>
				<td ref="number" onClick={this.handleClick}>{number}</td>
				<td ref="description" onClick={this.handleClick}>{this.props.transaction.Description}</td>
				<td ref="account" onClick={this.handleClick}>{accountName}</td>
				<td ref="status" onClick={this.handleClick}>{status}</td>
				<td ref="amount" onClick={this.handleClick}>{amount}</td>
				<td>{balance}</td>
			</tr>);
	}
});

const AddEditTransactionModal = React.createClass({
	_getInitialState: function(props) {
		// Ensure we can edit this without screwing up other copies of it
		var t = props.transaction.deepCopy();
		return {transaction: t};
	},
	getInitialState: function() {
		 return this._getInitialState(this.props);
	},
	handleCancel: function() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	},
	handleDescriptionChange: function() {
		var transaction = this.state.transaction.deepCopy();
		transaction.Description = this.refs.description.getValue();
		this.setState({
			transaction: transaction
		});
	},
	handleDateChange: function(date, string) {
		if (date == null)
			return;
		var transaction = this.state.transaction.deepCopy();
		transaction.Date = date;
		this.setState({
			transaction: transaction
		});
	},
	handleStatusChange: function(status) {
		if (status.hasOwnProperty('StatusId')) {
			var transaction = this.state.transaction.deepCopy();
			transaction.Status = status.StatusId;
			this.setState({
				transaction: transaction
			});
		}
	},
	handleDeleteSplit: function(split) {
		var transaction = this.state.transaction.deepCopy();
		transaction.Splits.splice(split, 1);
		this.setState({
			transaction: transaction
		});
	},
	handleUpdateNumber: function(split) {
		var transaction = this.state.transaction.deepCopy();
		transaction.Splits[split].Number = this.refs['number-'+split].getValue();
		this.setState({
			transaction: transaction
		});
	},
	handleUpdateMemo: function(split) {
		var transaction = this.state.transaction.deepCopy();
		transaction.Splits[split].Memo = this.refs['memo-'+split].getValue();
		this.setState({
			transaction: transaction
		});
	},
	handleUpdateAccount: function(account, split) {
		var transaction = this.state.transaction.deepCopy();
		transaction.Splits[split].AccountId = account.AccountId;
		this.setState({
			transaction: transaction
		});
	},
	handleUpdateAmount: function(split) {
		var transaction = this.state.transaction.deepCopy();
		transaction.Splits[split].Amount = new Big(this.refs['amount-'+split].getValue());
		this.setState({
			transaction: transaction
		});
	},
	handleSubmit: function() {
		if (this.props.onSubmit != null)
			this.props.onSubmit(this.state.transaction);
	},
	handleDelete: function() {
		if (this.props.onDelete != null)
			this.props.onDelete(this.state.transaction);
	},
	componentWillReceiveProps: function(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this._getInitialState(nextProps));
		}
	},
	render: function() {
		var editing = this.props.transaction != null && this.props.transaction.isTransaction();
		var headerText = editing ? "Edit" : "Create New";
		var buttonText = editing ? "Save Changes" : "Create Transaction";
		var deleteButton = [];
		if (editing) {
			deleteButton = (
				<Button onClick={this.handleDelete} bsStyle="danger">Delete Transaction</Button>
		   );
		}

		splits = [];
		for (var i = 0; i < this.state.transaction.Splits.length; i++) {
			var self = this;
			var s = this.state.transaction.Splits[i];

			// Define all closures for calling split-updating functions
			var deleteSplitFn = (function() {
				var j = i;
				return function() {self.handleDeleteSplit(j);};
			})();
			var updateNumberFn = (function() {
				var j = i;
				return function() {self.handleUpdateNumber(j);};
			})();
			var updateMemoFn = (function() {
				var j = i;
				return function() {self.handleUpdateMemo(j);};
			})();
			var updateAccountFn = (function() {
				var j = i;
				return function(account) {self.handleUpdateAccount(account, j);};
			})();
			var updateAmountFn = (function() {
				var j = i;
				return function() {self.handleUpdateAmount(j);};
			})();

			var deleteSplitButton = [];
			if (this.state.transaction.Splits.length > 2) {
				deleteSplitButton = (
					<Col xs={1}><Button onClick={deleteSplitFn}
							bsStyle="danger">
							<Glyphicon glyph='trash' /></Button></Col>
				);
			}

			splits.push((
				<Row>
				<Col xs={1}><Input
					type="text"
					value={s.Number}
					onChange={updateNumberFn}
					ref={"number-"+i} /></Col>
				<Col xs={5}><Input
					type="text"
					value={s.Memo}
					onChange={updateMemoFn}
					ref={"memo-"+i} /></Col>
				<Col xs={3}><AccountCombobox
					accounts={this.props.accounts}
					account_map={this.props.account_map}
					value={s.AccountId}
					includeRoot={false}
					onSelect={updateAccountFn}
					ref={"account-"+i} /></Col>
				<Col xs={2}><Input type="text"
					value={s.Amount}
					onChange={updateAmountFn}
					ref={"amount-"+i} /></Col>
				{deleteSplitButton}
				</Row>
			));
		}

		return (
			<Modal show={this.props.show} onHide={this.handleCancel} bsSize="large">
				<Modal.Header closeButton>
					<Modal.Title>{headerText} Transaction</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<form onSubmit={this.handleSubmit}
						className="form-horizontal">
					<Input wrapperClassName="wrapper"
						label="Date"
						labelClassName="col-xs-2"
						wrapperClassName="col-xs-10">
					<DateTimePicker
						time={false}
						defaultValue={this.state.transaction.Date}
						onChange={this.handleDateChange} />
					</Input>
					<Input type="text"
						label="Description"
						value={this.state.transaction.Description}
						onChange={this.handleDescriptionChange}
						ref="description"
						labelClassName="col-xs-2"
						wrapperClassName="col-xs-10"/>
					<Input wrapperClassName="wrapper"
						label="Status"
						labelClassName="col-xs-2"
						wrapperClassName="col-xs-10">
					<Combobox
						data={TransactionStatusList}
						valueField='StatusId'
						textField='Name'
						value={this.state.transaction.Status}
						onSelect={this.handleStatusChange}
						ref="status" />
					</Input>
					<Grid fluid={true}><Row>
					<span className="split-header col-xs-1">#</span>
					<span className="split-header col-xs-5">Memo</span>
					<span className="split-header col-xs-3">Account</span>
					<span className="split-header col-xs-2">Amount</span>
					</Row>
					{splits}
					</Grid>
				</form>
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup>
						<Button onClick={this.handleCancel} bsStyle="warning">Cancel</Button>
						{deleteButton}
						<Button onClick={this.handleSubmit} bsStyle="success">{buttonText}</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
});

const AccountRegister = React.createClass({
	getInitialState: function() {
		return {
			editingTransaction: false,
			selectedTransaction: new Transaction(),
			transactions: [],
			pageSize: 20,
			numPages: 0,
			currentPage: 0,
			height: 0
		};
	},
	resize: function() {
		var div = React.findDOMNode(this);
		this.setState({height: div.parentElement.clientHeight - 64});
	},
	componentDidMount: function() {
		this.resize();
		var self = this;
		$(window).resize(function() {self.resize();});
	},
	handleEditTransaction: function(transaction) {
		this.setState({
			selectedTransaction: transaction,
			editingTransaction: true
		});
	},
	handleEditingCancel: function() {
		this.setState({
			editingTransaction: false
		});
	},
	handleNewTransactionClicked: function() {
		var newTransaction = new Transaction();
		newTransaction.Status = TransactionStatus.Entered;
		newTransaction.Date = new Date();
		newTransaction.Splits.push(new Split());
		newTransaction.Splits.push(new Split());
		newTransaction.Splits[0].AccountId = this.props.selectedAccount.AccountId;

		this.setState({
			editingTransaction: true,
			selectedTransaction: newTransaction
		});
	},
	ajaxError: function(jqXHR, status, error) {
		var e = new Error();
		e.ErrorId = 5;
		e.ErrorString = "Request Failed: " + status + error;
		this.setState({error: e});
	},
	getTransactionPage: function(account, page) {
		$.ajax({
			type: "GET",
			dataType: "json",
			url: "account/"+account.AccountId+"/transactions?sort=date-desc&limit="+this.state.pageSize+"&page="+page,
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
					return;
				}

				var transactions = [];
				var balance = new Big(data.BeginningBalance);

				for (var i = 0; i < data.Transactions.length; i++) {
					var t = new Transaction();
					t.fromJSON(data.Transactions[i]);

					// Keep a talley of the running balance of these transactions
					for (var j = 0; j < data.Transactions[i].Splits.length; j++) {
						var split = data.Transactions[i].Splits[j];
						if (this.props.selectedAccount.AccountId == split.AccountId) {
							balance = balance.plus(split.Amount);
						}
					}
					t.Balance = balance.plus(0); // Make a copy
					transactions.push(t);
				}
				var a = new Account();
				a.fromJSON(data.Account);

				var pages = Math.ceil(data.TotalTransactions / this.state.pageSize);

				this.setState({
					transactions: transactions,
					numPages: pages
				});
			}.bind(this),
			error: this.ajaxError
		});
	},
	handleSelectPage: function(event, selectedEvent) {
		var newpage = selectedEvent.eventKey - 1;
		// Don't do pages that don't make sense
		if (newpage < 0)
			newpage = 0;
		if (newpage >= this.state.numPages)
			newpage = this.state.numPages-1;
		if (newpage != this.state.currentPage) {
			if (this.props.selectedAccount != null) {
				this.getTransactionPage(this.props.selectedAccount, newpage);
			}
			this.setState({currentPage: newpage});
		}
	},
	onNewTransaction: function() {
		this.getTransactionPage(this.props.selectedAccount, this.state.currentPage);
	},
	onUpdatedTransaction: function() {
		this.getTransactionPage(this.props.selectedAccount, this.state.currentPage);
	},
	onDeletedTransaction: function() {
		this.getTransactionPage(this.props.selectedAccount, this.state.currentPage);
	},
	createNewTransaction: function(transaction) {
		$.ajax({
			type: "POST",
			dataType: "json",
			url: "transaction/",
			data: {transaction: transaction.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					this.onNewTransaction();
				}
			}.bind(this),
			error: this.ajaxError
		});
	},
	updateTransaction: function(transaction) {
		$.ajax({
			type: "PUT",
			dataType: "json",
			url: "transaction/"+transaction.TransactionId+"/",
			data: {transaction: transaction.toJSON()},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					this.onUpdatedTransaction();
				}
			}.bind(this),
			error: this.ajaxError
		});
	},
	deleteTransaction: function(transaction) {
		console.log("handleDeleteTransaction", transaction);
		$.ajax({
			type: "DELETE",
			dataType: "json",
			url: "transaction/"+transaction.TransactionId+"/",
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					this.setState({error: e});
				} else {
					this.onDeletedTransaction();
				}
			}.bind(this),
			error: this.ajaxError
		});
	},
	handleDeleteTransaction: function(transaction) {
		this.setState({
			editingTransaction: false
		});
		this.deleteTransaction(transaction);
	},
	handleUpdateTransaction: function(transaction) {
		this.setState({
			editingTransaction: false
		});
		if (transaction.TransactionId != -1) {
			this.updateTransaction(transaction);
		} else {
			this.createNewTransaction(transaction);
		}
	},
	componentWillReceiveProps: function(nextProps) {
		if (nextProps.selectedAccount != this.props.selectedAccount) {
			this.setState({
				selectedTransaction: new Transaction(),
				transactions: [],
				currentPage: 0
			});
			this.getTransactionPage(nextProps.selectedAccount, 0);
		}
	},
	render: function() {
		var name = "Please select an account";
		register = [];

		if (this.props.selectedAccount != null) {
			name = this.props.selectedAccount.Name;

			var transactionRows = [];
			for (var i = 0; i < this.state.transactions.length; i++) {
				var t = this.state.transactions[i];
				transactionRows.push((
					<TransactionRow
						transaction={t}
						account={this.props.selectedAccount}
						accounts={this.props.accounts}
						account_map={this.props.account_map}
						securities={this.props.securities}
						security_map={this.props.security_map}
						onEdit={this.handleEditTransaction}/>
				));
			}

			var style = {height: this.state.height + "px"};
			register = (
				<div style={style} className="transactions-register">
				<Table bordered striped condensed hover>
					<thead><tr>
						<th>Date</th>
						<th>#</th>
						<th>Description</th>
						<th>Account</th>
						<th>Status</th>
						<th>Amount</th>
						<th>Balance</th>
					</tr></thead>
					<tbody>
						{transactionRows}
					</tbody>
				</Table>
				</div>
			);
		}

		var disabled = (this.props.selectedAccount == null) ? "disabled" : "";

		return (
			<div className="transactions-container">
				<AddEditTransactionModal
					show={this.state.editingTransaction}
					transaction={this.state.selectedTransaction}
					accounts={this.props.accounts}
					account_map={this.props.account_map}
					onCancel={this.handleEditingCancel}
					onSubmit={this.handleUpdateTransaction}
					onDelete={this.handleDeleteTransaction}
					securities={this.props.securities}/>
				<div className="transactions-register-toolbar">
				Transactions for '{name}'
				<ButtonToolbar className="pull-right">
					<ButtonGroup>
					<Pagination
						className="skinny-pagination"
						prev next first last ellipses
						items={this.state.numPages}
						maxButtons={Math.min(5, this.state.numPages)}
						activePage={this.state.currentPage+1}
						onSelect={this.handleSelectPage} />
					</ButtonGroup>
					<ButtonGroup>
					<Button
							onClick={this.handleNewTransactionClicked}
							bsStyle="success"
							disabled={disabled}>
						<Glyphicon glyph='plus-sign' /> New Transaction
					</Button>
					</ButtonGroup>
				</ButtonToolbar>
				</div>
				{register}
			</div>
		);
	}
});
