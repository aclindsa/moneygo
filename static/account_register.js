// Import all the objects we want to use from ReactBootstrap

var Alert = ReactBootstrap.Alert;
var Modal = ReactBootstrap.Modal;
var Pagination = ReactBootstrap.Pagination;

var Label = ReactBootstrap.Label;
var Table = ReactBootstrap.Table;
var Grid = ReactBootstrap.Grid;
var Row = ReactBootstrap.Row;
var Col = ReactBootstrap.Col;
var Panel = ReactBootstrap.Panel;

var Button = ReactBootstrap.Button;
var ButtonToolbar = ReactBootstrap.ButtonToolbar;

var ProgressBar = ReactBootstrap.ProgressBar;

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

				if (otherSplit.AccountId == -1)
					var accountName = "Unbalanced " + this.props.security_map[otherSplit.SecurityId].Symbol + " transaction";
				else
					var accountName = getAccountDisplayName(this.props.account_map[otherSplit.AccountId], this.props.account_map);
			} else {
				accountName = "--Split Transaction--";
			}

			var amount = security.Symbol + " " + thisAccountSplit.Amount.toFixed(security.Precision);
			var balance = security.Symbol + " " + this.props.transaction.Balance.toFixed(security.Precision);
			status = TransactionStatusMap[this.props.transaction.Status];
			number = thisAccountSplit.Number;
		} else {
			var amount = security.Symbol + " " + (new Big(0.0)).toFixed(security.Precision);
			var balance = security.Symbol + " " + (new Big(0.0)).toFixed(security.Precision);
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

const AmountInput = React.createClass({
	_getInitialState: function(props) {
		// Ensure we can edit this without screwing up other copies of it
		var a;
		if (props.security)
			a = props.value.toFixed(props.security.Precision);
		else
			a = props.value.toString();

		return {
			LastGoodAmount: a,
			Amount: a
		};
	},
	getInitialState: function() {
		 return this._getInitialState(this.props);
	},
	componentWillReceiveProps: function(nextProps) {
		if ((!nextProps.value.eq(this.props.value) &&
				!nextProps.value.eq(this.getValue())) ||
				nextProps.security !== this.props.security) {
			this.setState(this._getInitialState(nextProps));
		}
	},
	componentDidMount: function() {
		this.refs.amount.getInputDOMNode().onblur = this.onBlur;
	},
	onBlur: function() {
		var a;
		if (this.props.security)
			a = (new Big(this.getValue())).toFixed(this.props.security.Precision);
		else
			a = (new Big(this.getValue())).toString();
		this.setState({
			Amount: a
		});
	},
	onChange: function() {
		this.setState({Amount: this.refs.amount.getValue()});
		if (this.props.onChange)
			this.props.onChange();
	},
	getValue: function() {
		try {
			var value = this.refs.amount.getValue();
			var ret = new Big(value);
			this.setState({LastGoodAmount: value});
			return ret;
		} catch(err) {
			return new Big(this.state.LastGoodAmount);
		}
	},
	render: function() {
		var symbol = "?";
		if (this.props.security)
			symbol = this.props.security.Symbol;
		var bsStyle = "";
		if (this.props.bsStyle)
			bsStyle = this.props.bsStyle;

		return (
			<Input type="text"
				value={this.state.Amount}
				onChange={this.onChange}
				addonBefore={symbol}
				bsStyle={bsStyle}
				ref="amount"/>
		);
	}
});

const AddEditTransactionModal = React.createClass({
	_getInitialState: function(props) {
		// Ensure we can edit this without screwing up other copies of it
		var t = props.transaction.deepCopy();
		return {
			errorAlert: [],
			transaction: t
		};
	},
	getInitialState: function() {
		 return this._getInitialState(this.props);
	},
	componentWillReceiveProps: function(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this._getInitialState(nextProps));
		}
	},
	handleCancel: function() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	},
	handleDescriptionChange: function() {
		this.setState({
			transaction: React.addons.update(this.state.transaction, {
				Description: {$set: this.refs.description.getValue()}
			})
		});
	},
	handleDateChange: function(date, string) {
		if (date == null)
			return;
		this.setState({
			transaction: React.addons.update(this.state.transaction, {
				Date: {$set: date}
			})
		});
	},
	handleStatusChange: function(status) {
		if (status.hasOwnProperty('StatusId')) {
			this.setState({
				transaction: React.addons.update(this.state.transaction, {
					Status: {$set: status.StatusId}
				})
			});
		}
	},
	handleAddSplit: function() {
		this.setState({
			transaction: React.addons.update(this.state.transaction, {
				Splits: {$push: [new Split()]}
			})
		});
	},
	handleDeleteSplit: function(split) {
		this.setState({
			transaction: React.addons.update(this.state.transaction, {
				Splits: {$splice: [[split, 1]]}
			})
		});
	},
	handleUpdateNumber: function(split) {
		var transaction = this.state.transaction;
		transaction.Splits[split] = React.addons.update(transaction.Splits[split], {
			Number: {$set: this.refs['number-'+split].getValue()}
		});
		this.setState({
			transaction: transaction
		});
	},
	handleUpdateMemo: function(split) {
		var transaction = this.state.transaction;
		transaction.Splits[split] = React.addons.update(transaction.Splits[split], {
			Memo: {$set: this.refs['memo-'+split].getValue()}
		});
		this.setState({
			transaction: transaction
		});
	},
	handleUpdateAccount: function(account, split) {
		var transaction = this.state.transaction;
		transaction.Splits[split] = React.addons.update(transaction.Splits[split], {
			SecurityId: {$set: -1},
			AccountId: {$set: account.AccountId}
		});
		this.setState({
			transaction: transaction
		});
	},
	handleUpdateAmount: function(split) {
		var transaction = this.state.transaction;
		transaction.Splits[split] = React.addons.update(transaction.Splits[split], {
			Amount: {$set: new Big(this.refs['amount-'+split].getValue())}
		});
		this.setState({
			transaction: transaction
		});
	},
	handleSubmit: function() {
		var errorString = ""
		var imbalancedSecurityList = this.state.transaction.imbalancedSplitSecurities(this.props.account_map);
		if (imbalancedSecurityList.length > 0)
			errorString = "Transaction must balance"
		for (var i = 0; i < this.state.transaction.Splits.length; i++) {
			var s = this.state.transaction.Splits[i];
			if (!(s.AccountId in this.props.account_map)) {
				errorString = "All accounts must be valid"
			}
		}

		if (errorString.length > 0) {
			this.setState({
				errorAlert: (<Alert className='saving-transaction-alert' bsStyle='danger'><strong>Error Saving Transaction:</strong> {errorString}</Alert>)
			});
			return;
		}

		if (this.props.onSubmit != null)
			this.props.onSubmit(this.state.transaction);
	},
	handleDelete: function() {
		if (this.props.onDelete != null)
			this.props.onDelete(this.state.transaction);
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

		var imbalancedSecurityList = this.state.transaction.imbalancedSplitSecurities(this.props.account_map);
		var imbalancedSecurityMap = {};
		for (i = 0; i < imbalancedSecurityList.length; i++)
			imbalancedSecurityMap[imbalancedSecurityList[i]] = i;

		splits = [];
		for (var i = 0; i < this.state.transaction.Splits.length; i++) {
			var self = this;
			var s = this.state.transaction.Splits[i];
			var security = null;
			var amountValidation = "";
			var accountValidation = "";
			if (s.AccountId in this.props.account_map) {
				security = this.props.security_map[this.props.account_map[s.AccountId].SecurityId];
			} else {
				if (s.SecurityId in this.props.security_map) {
					security = this.props.security_map[s.SecurityId];
				}
				accountValidation = "has-error";
			}
			if (security != null && security.SecurityId in imbalancedSecurityMap)
				amountValidation = "error";

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
					ref={"account-"+i}
					className={accountValidation}/></Col>
				<Col xs={2}><AmountInput type="text"
					value={s.Amount}
					security={security}
					onChange={updateAmountFn}
					ref={"amount-"+i}
					bsStyle={amountValidation}/></Col>
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
					<Row>
						<span className="col-xs-11"></span>
						<Col xs={1}><Button onClick={this.handleAddSplit}
								bsStyle="success">
								<Glyphicon glyph='plus-sign' /></Button></Col>
					</Row>
					<Row>{this.state.errorAlert}</Row>
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

const ImportTransactionsModal = React.createClass({
	getInitialState: function() {
		 return {
			importing: false,
			imported: false,
			importFile: "",
			uploadProgress: -1,
			error: null};
	},
	handleCancel: function() {
		this.setState({
			importing: false,
			imported: false,
			importFile: "",
			uploadProgress: -1,
			error: null
		});
		if (this.props.onCancel != null)
			this.props.onCancel();
	},
	onImportChanged: function() {
		this.setState({importFile: this.refs.importfile.getValue()});
	},
	handleSubmit: function() {
		if (this.props.onSubmit != null)
			this.props.onSubmit(this.props.account);
	},
	handleSetProgress: function(e) {
		if (e.lengthComputable) {
			var pct = Math.round(e.loaded/e.total*100);
			this.setState({uploadProgress: pct});
		} else {
			this.setState({uploadProgress: 50});
		}
	},
	handleImportTransactions: function() {
		var file = this.refs.importfile.getInputDOMNode().files[0];
		var formData = new FormData();
		this.setState({importing: true});
		formData.append('importfile', file, this.state.importFile);
		$.ajax({
			type: "POST",
			url: "account/"+this.props.account.AccountId+"/import",
			data: formData,
			xhr: function() {
				var xhrObject = $.ajaxSettings.xhr();
				if (xhrObject.upload) {
					xhrObject.upload.addEventListener('progress', this.handleSetProgress, false);
				} else {
					console.log("File upload failed because !xhr.upload")
				}
				return xhrObject;
			}.bind(this),
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					var errString = e.ErrorString;
					if (e.ErrorId == 3 /* Invalid Request */) {
						errString = "Please check that the file you uploaded is a valid OFX file for this account and try again.";
					}
					this.setState({
						importing: false,
						error: errString
					});
					return;
				}

				this.setState({
					uploadProgress: 100,
					importing: false,
					imported: true
				});
			}.bind(this),
			error: function(e) {
				this.setState({importing: false});
				console.log("error handler", e);
			},
			// So jQuery doesn't try to process teh data or content-type
			cache: false,
			contentType: false,
			processData: false
		});
	},
	render: function() {
		var accountNameLabel = ""
		if (this.props.account != null )
			accountNameLabel = "Importing to '" + getAccountDisplayName(this.props.account, this.props.account_map) + "' account:";
		var progressBar = [];
		if (this.state.importing && this.state.uploadProgress == 100) {
			progressBar = (<ProgressBar now={this.state.uploadProgress} active label="Importing transactions..." />);
		} else if (this.state.importing && this.state.uploadProgress != -1) {
			progressBar = (<ProgressBar now={this.state.uploadProgress} active label="Uploading... %(percent)s%" />);
		}

		var panel = [];
		if (this.state.error != null) {
			panel = (<Panel header="Error Importing Transactions" bsStyle="danger">{this.state.error}</Panel>);
		} else if (this.state.imported) {
			panel = (<Panel header="Successfully Imported Transactions" bsStyle="success">Your import is now complete.</Panel>);
		}

		var buttonsDisabled = (this.state.importing) ? "disabled" : "";
		var button1 = [];
		var button2 = [];
		if (!this.state.imported && this.state.error == null) {
			button1 = (<Button onClick={this.handleCancel} disabled={buttonsDisabled} bsStyle="warning">Cancel</Button>);
			button2 = (<Button onClick={this.handleImportTransactions} disabled={buttonsDisabled} bsStyle="success">Import</Button>);
		} else {
			button1 = (<Button onClick={this.handleCancel} disabled={buttonsDisabled} bsStyle="success">OK</Button>);
		}
		var inputDisabled = (this.state.importing || this.state.error != null || this.state.imported) ? "disabled" : "";
		return (
			<Modal show={this.props.show} onHide={this.handleCancel} bsSize="medium">
				<Modal.Header closeButton>
					<Modal.Title>Import Transactions</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<form onSubmit={this.handleImportTransactions}
						encType="multipart/form-data"
						ref="importform">
					<Input type="file"
							ref="importfile"
							disabled={inputDisabled}
							value={this.state.importFile}
							label={accountNameLabel}
							help="Select an OFX/QFX file to upload."
							onChange={this.onImportChanged} />
				</form>
				{progressBar}
				{panel}
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup>
						{button1}
						{button2}
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
});

const AccountRegister = React.createClass({
	getInitialState: function() {
		return {
			importingTransactions: false,
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
	handleImportClicked: function() {
		this.setState({
			importingTransactions: true
		});
	},
	handleImportingCancel: function() {
		this.setState({
			importingTransactions: false
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
				var balance = new Big(data.EndingBalance);

				for (var i = 0; i < data.Transactions.length; i++) {
					var t = new Transaction();
					t.fromJSON(data.Transactions[i]);

					t.Balance = balance.plus(0); // Make a copy of the current balance
					// Keep a talley of the running balance of these transactions
					for (var j = 0; j < data.Transactions[i].Splits.length; j++) {
						var split = data.Transactions[i].Splits[j];
						if (this.props.selectedAccount.AccountId == split.AccountId) {
							balance = balance.minus(split.Amount);
						}
					}
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
	handleImportComplete: function() {
		this.setState({importingTransactions: false});
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
					securities={this.props.securities}
					security_map={this.props.security_map}/>
				<ImportTransactionsModal
					show={this.state.importingTransactions}
					account={this.props.selectedAccount}
					accounts={this.props.accounts}
					account_map={this.props.account_map}
					onCancel={this.handleImportingCancel}
					onSubmit={this.handleImportComplete}/>
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
					<Button
							onClick={this.handleImportClicked}
							bsStyle="primary"
							disabled={disabled}>
						<Glyphicon glyph='import' /> Import
					</Button>
					</ButtonGroup>
				</ButtonToolbar>
				</div>
				{register}
			</div>
		);
	}
});
