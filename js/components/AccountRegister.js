var React = require('react');
var ReactDOM = require('react-dom');

var react_update = require('react-addons-update');

var ReactBootstrap = require('react-bootstrap');
var Alert = ReactBootstrap.Alert;
var Modal = ReactBootstrap.Modal;
var Pagination = ReactBootstrap.Pagination;
var Label = ReactBootstrap.Label;
var Table = ReactBootstrap.Table;
var Grid = ReactBootstrap.Grid;
var Row = ReactBootstrap.Row;
var Col = ReactBootstrap.Col;
var Panel = ReactBootstrap.Panel;
var Form = ReactBootstrap.Form;
var FormGroup = ReactBootstrap.FormGroup;
var FormControl = ReactBootstrap.FormControl;
var InputGroup = ReactBootstrap.InputGroup;
var ControlLabel = ReactBootstrap.ControlLabel;
var HelpBlock = ReactBootstrap.HelpBlock;
var Button = ReactBootstrap.Button;
var ButtonGroup = ReactBootstrap.ButtonGroup;
var ButtonToolbar = ReactBootstrap.ButtonToolbar;
var ProgressBar = ReactBootstrap.ProgressBar;
var Glyphicon = ReactBootstrap.Glyphicon;

var ReactWidgets = require('react-widgets')
var DateTimePicker = ReactWidgets.DateTimePicker;
var Combobox = ReactWidgets.Combobox;
var DropdownList = ReactWidgets.DropdownList;

var Big = require('big.js');

var models = require('../models');
var Security = models.Security;
var Account = models.Account;
var SplitStatus = models.SplitStatus;
var SplitStatusList = models.SplitStatusList;
var SplitStatusMap = models.SplitStatusMap;
var Split = models.Split;
var Transaction = models.Transaction;
var Error = models.Error;

var getAccountDisplayName = require('../utils').getAccountDisplayName;

var AccountCombobox = require('./AccountCombobox');

class TransactionRow extends React.Component {
	constructor() {
		super();
		this.onClick = this.handleClick.bind(this);
	}
	handleClick(e) {
		const refs = ["date", "number", "description", "account", "status", "amount"];
		for (var ref in refs) {
			if (this.refs[refs[ref]] == e.target) {
				this.props.onSelect(this.props.transaction.TransactionId, refs[ref]);
				return;
			}
		}
	}
	render() {
		var date = this.props.transaction.Date;
		var dateString = date.getFullYear() + "/" + (date.getMonth()+1) + "/" + date.getDate();
		var number = ""
		var accountName = "";
		var status = "";
		var security = this.props.securities[this.props.account.SecurityId];
		var balance = security.Symbol + " " + "?"

		if (this.props.transaction.isTransaction()) {
			var thisAccountSplit;
			for (var i = 0; i < this.props.transaction.Splits.length; i++) {
				if (this.props.transaction.Splits[i].AccountId == this.props.account.AccountId) {
					thisAccountSplit = this.props.transaction.Splits[i];
					status = SplitStatusMap[this.props.transaction.Splits[i].Status];
					break;
				}
			}
			if (this.props.transaction.Splits.length == 2) {
				var otherSplit = this.props.transaction.Splits[0];
				if (otherSplit.AccountId == this.props.account.AccountId)
					var otherSplit = this.props.transaction.Splits[1];

				if (otherSplit.AccountId == -1)
					var accountName = "Unbalanced " + this.props.securities[otherSplit.SecurityId].Symbol + " transaction";
				else
					var accountName = getAccountDisplayName(this.props.accounts[otherSplit.AccountId], this.props.accounts);
			} else {
				accountName = "--Split Transaction--";
			}

			var amount = security.Symbol + " " + thisAccountSplit.Amount.toFixed(security.Precision);
			if (this.props.transaction.hasOwnProperty("Balance"))
				balance = security.Symbol + " " + this.props.transaction.Balance.toFixed(security.Precision);
			number = thisAccountSplit.Number;
		} else {
			var amount = security.Symbol + " " + (new Big(0.0)).toFixed(security.Precision);
		}

		return (
			<tr>
				<td ref="date" onClick={this.onClick}>{dateString}</td>
				<td ref="number" onClick={this.onClick}>{number}</td>
				<td ref="description" onClick={this.onClick}>{this.props.transaction.Description}</td>
				<td ref="account" onClick={this.onClick}>{accountName}</td>
				<td ref="status" onClick={this.onClick}>{status}</td>
				<td ref="amount" onClick={this.onClick}>{amount}</td>
				<td>{balance}</td>
			</tr>);
	}
}

class AmountInput extends React.Component {
	getInitialState(props) {
		if (!props)
			return {
				LastGoodAmount: "0",
				Amount: "0"
			}

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
	}
	constructor() {
		super();
		this.onChange = this.handleChange.bind(this);
		this.state = this.getInitialState();
	}
	componentWillReceiveProps(nextProps) {
		if ((!nextProps.value.eq(this.props.value) &&
				!nextProps.value.eq(this.getValue())) ||
				nextProps.security !== this.props.security) {
			this.setState(this.getInitialState(nextProps));
		}
	}
	componentDidMount() {
		ReactDOM.findDOMNode(this.refs.amount).onblur = this.handleBlur.bind(this);
	}
	handleBlur() {
		var a;
		if (this.props.security)
			a = (new Big(this.getValue())).toFixed(this.props.security.Precision);
		else
			a = (new Big(this.getValue())).toString();
		this.setState({
			Amount: a
		});
	}
	handleChange() {
		this.setState({Amount: ReactDOM.findDOMNode(this.refs.amount).value});
		if (this.props.onChange)
			this.props.onChange();
	}
	getValue() {
		try {
			var value = ReactDOM.findDOMNode(this.refs.amount).value;
			var ret = new Big(value);
			this.setState({LastGoodAmount: value});
			return ret;
		} catch(err) {
			return new Big(this.state.LastGoodAmount);
		}
	}
	render() {
		var symbol = "?";
		if (this.props.security)
			symbol = this.props.security.Symbol;

		return (
			<FormGroup validationState={this.props.validationState}>
				<InputGroup>
					<InputGroup.Addon>{symbol}</InputGroup.Addon>
					<FormControl type="text"
						value={this.state.Amount}
						onChange={this.onChange}
						ref="amount"/>
				</InputGroup>
			</FormGroup>
		);
	}
}

class AddEditTransactionModal extends React.Component {
	getInitialState(props) {
		// Ensure we can edit this without screwing up other copies of it
		if (props)
			var t = props.transaction.deepCopy();
		else
			var t = new Transaction();

		return {
			errorAlert: [],
			transaction: t
		};
	}
	constructor() {
		super();
		this.state = this.getInitialState();
		this.onCancel = this.handleCancel.bind(this);
		this.onDescriptionChange = this.handleDescriptionChange.bind(this);
		this.onDateChange = this.handleDateChange.bind(this);
		this.onAddSplit = this.handleAddSplit.bind(this);
		this.onDeleteSplit = this.handleDeleteSplit.bind(this);
		this.onUpdateNumber = this.handleUpdateNumber.bind(this);
		this.onUpdateStatus = this.handleUpdateStatus.bind(this);
		this.onUpdateMemo = this.handleUpdateMemo.bind(this);
		this.onUpdateAccount = this.handleUpdateAccount.bind(this);
		this.onUpdateAmount = this.handleUpdateAmount.bind(this);
		this.onSubmit = this.handleSubmit.bind(this);
		this.onDelete = this.handleDelete.bind(this);
	}
	componentWillReceiveProps(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this.getInitialState(nextProps));
		}
	}
	handleCancel() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	}
	handleDescriptionChange() {
		this.setState({
			transaction: react_update(this.state.transaction, {
				Description: {$set: ReactDOM.findDOMNode(this.refs.description).value}
			})
		});
	}
	handleDateChange(date, string) {
		if (date == null)
			return;
		this.setState({
			transaction: react_update(this.state.transaction, {
				Date: {$set: date}
			})
		});
	}
	handleAddSplit() {
		var split = new Split();
		split.Status = SplitStatus.Entered;
		this.setState({
			transaction: react_update(this.state.transaction, {
				Splits: {$push: [split]}
			})
		});
	}
	handleDeleteSplit(split) {
		this.setState({
			transaction: react_update(this.state.transaction, {
				Splits: {$splice: [[split, 1]]}
			})
		});
	}
	handleUpdateNumber(split) {
		var transaction = this.state.transaction;
		transaction.Splits[split] = react_update(transaction.Splits[split], {
			Number: {$set: ReactDOM.findDOMNode(this.refs['number-'+split]).value}
		});
		this.setState({
			transaction: transaction
		});
	}
	handleUpdateStatus(status, split) {
		var transaction = this.state.transaction;
		transaction.Splits[split] = react_update(transaction.Splits[split], {
			Status: {$set: status.StatusId}
		});
		this.setState({
			transaction: transaction
		});
	}
	handleUpdateMemo(split) {
		var transaction = this.state.transaction;
		transaction.Splits[split] = react_update(transaction.Splits[split], {
			Memo: {$set: ReactDOM.findDOMNode(this.refs['memo-'+split]).value}
		});
		this.setState({
			transaction: transaction
		});
	}
	handleUpdateAccount(account, split) {
		var transaction = this.state.transaction;
		transaction.Splits[split] = react_update(transaction.Splits[split], {
			SecurityId: {$set: -1},
			AccountId: {$set: account.AccountId}
		});
		this.setState({
			transaction: transaction
		});
	}
	handleUpdateAmount(split) {
		var transaction = this.state.transaction;
		transaction.Splits[split] = react_update(transaction.Splits[split], {
			Amount: {$set: new Big(this.refs['amount-'+split].getValue())}
		});
		this.setState({
			transaction: transaction
		});
	}
	handleSubmit() {
		var errorString = ""
		var imbalancedSecurityList = this.state.transaction.imbalancedSplitSecurities(this.props.accounts);
		if (imbalancedSecurityList.length > 0)
			errorString = "Transaction must balance"
		for (var i = 0; i < this.state.transaction.Splits.length; i++) {
			var s = this.state.transaction.Splits[i];
			if (!(s.AccountId in this.props.accounts)) {
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
	}
	handleDelete() {
		if (this.props.onDelete != null)
			this.props.onDelete(this.state.transaction);
	}
	render() {
		var editing = this.props.transaction != null && this.props.transaction.isTransaction();
		var headerText = editing ? "Edit" : "Create New";
		var buttonText = editing ? "Save Changes" : "Create Transaction";
		var deleteButton = [];
		if (editing) {
			deleteButton = (
				<Button key={1} onClick={this.onDelete} bsStyle="danger">Delete Transaction</Button>
		   );
		}

		var imbalancedSecurityList = this.state.transaction.imbalancedSplitSecurities(this.props.accounts);
		var imbalancedSecurityMap = {};
		for (i = 0; i < imbalancedSecurityList.length; i++)
			imbalancedSecurityMap[imbalancedSecurityList[i]] = i;

		var splits = [];
		for (var i = 0; i < this.state.transaction.Splits.length; i++) {
			var self = this;
			var s = this.state.transaction.Splits[i];
			var security = null;
			var amountValidation = undefined;
			var accountValidation = "";
			if (s.AccountId in this.props.accounts) {
				security = this.props.securities[this.props.accounts[s.AccountId].SecurityId];
			} else {
				if (s.SecurityId in this.props.securities) {
					security = this.props.securities[s.SecurityId];
				}
				accountValidation = "has-error";
			}
			if (security != null && security.SecurityId in imbalancedSecurityMap)
				amountValidation = "error";

			// Define all closures for calling split-updating functions
			var deleteSplitFn = (function() {
				var j = i;
				return function() {self.onDeleteSplit(j);};
			})();
			var updateNumberFn = (function() {
				var j = i;
				return function() {self.onUpdateNumber(j);};
			})();
			var updateStatusFn = (function() {
				var j = i;
				return function(status) {self.onUpdateStatus(status, j);};
			})();
			var updateMemoFn = (function() {
				var j = i;
				return function() {self.onUpdateMemo(j);};
			})();
			var updateAccountFn = (function() {
				var j = i;
				return function(account) {self.onUpdateAccount(account, j);};
			})();
			var updateAmountFn = (function() {
				var j = i;
				return function() {self.onUpdateAmount(j);};
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
				<Row key={s.SplitId == -1 ? (i+999) : s.SplitId}>
				<Col xs={1}><FormControl
					type="text"
					value={s.Number}
					onChange={updateNumberFn}
					ref={"number-"+i} /></Col>
				<Col xs={1}>
					<Combobox
						suggest
						data={SplitStatusList}
						valueField='StatusId'
						textField='Name'
						defaultValue={s.Status}
						onSelect={updateStatusFn}
						ref={"status-"+i} />
					</Col>
				<Col xs={4}><FormControl
					type="text"
					value={s.Memo}
					onChange={updateMemoFn}
					ref={"memo-"+i} /></Col>
				<Col xs={3}><AccountCombobox
					accounts={this.props.accounts}
					accountChildren={this.props.accountChildren}
					value={s.AccountId}
					includeRoot={false}
					onChange={updateAccountFn}
					ref={"account-"+i}
					className={accountValidation}/></Col>
				<Col xs={2}><AmountInput type="text"
					value={s.Amount}
					security={security}
					onChange={updateAmountFn}
					ref={"amount-"+i}
					validationState={amountValidation}/></Col>
				{deleteSplitButton}
				</Row>
			));
		}

		return (
			<Modal show={this.props.show} onHide={this.onCancel} bsSize="large">
				<Modal.Header closeButton>
					<Modal.Title>{headerText} Transaction</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<Form horizontal
						onSubmit={this.onSubmit}>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Date</Col>
						<Col xs={10}>
						<DateTimePicker
							time={false}
							defaultValue={this.state.transaction.Date}
							onChange={this.onDateChange} />
						</Col>
					</FormGroup>
					<FormGroup>
						<Col componentClass={ControlLabel} xs={2}>Description</Col>
						<Col xs={10}>
						<FormControl type="text"
							value={this.state.transaction.Description}
							onChange={this.onDescriptionChange}
							ref="description"/>
						</Col>
					</FormGroup>

					<Grid fluid={true}><Row>
					<span className="split-header col-xs-1">#</span>
					<span className="split-header col-xs-1">Status</span>
					<span className="split-header col-xs-4">Memo</span>
					<span className="split-header col-xs-3">Account</span>
					<span className="split-header col-xs-2">Amount</span>
					</Row>
					{splits}
					<Row>
						<span className="col-xs-11"></span>
						<Col xs={1}><Button onClick={this.onAddSplit}
								bsStyle="success">
								<Glyphicon glyph='plus-sign' /></Button></Col>
					</Row>
					<Row>{this.state.errorAlert}</Row>
					</Grid>
				</Form>
				</Modal.Body>
				<Modal.Footer>
					<ButtonGroup>
						<Button onClick={this.onCancel} bsStyle="warning">Cancel</Button>
						{deleteButton}
						<Button onClick={this.onSubmit} bsStyle="success">{buttonText}</Button>
					</ButtonGroup>
				</Modal.Footer>
			</Modal>
		);
	}
}

const ImportType = {
	OFX: 1,
	OFXFile: 2,
	Gnucash: 3
};
var ImportTypeList = [];
for (var type in ImportType) {
	if (ImportType.hasOwnProperty(type)) {
		var name = ImportType[type] == ImportType.OFX ? "Direct OFX" : type;
		var name = ImportType[type] == ImportType.OFXFile ? "OFX/QFX File" : type; //QFX is a special snowflake
		ImportTypeList.push({'TypeId': ImportType[type], 'Name': name});
   }
}

class ImportTransactionsModal extends React.Component {
	getInitialState() {
		var startDate = new Date();
		startDate.setMonth(startDate.getMonth() - 1);
		return {
			importFile: "",
			importType: ImportType.OFX,
			startDate: startDate,
			endDate: new Date(),
			password: "",
		};
	}
	constructor() {
		super();
		this.state = this.getInitialState();
		this.onCancel = this.handleCancel.bind(this);
		this.onImportChange = this.handleImportChange.bind(this);
		this.onTypeChange = this.handleTypeChange.bind(this);
		this.onPasswordChange = this.handlePasswordChange.bind(this);
		this.onStartDateChange = this.handleStartDateChange.bind(this);
		this.onEndDateChange = this.handleEndDateChange.bind(this);
		this.onSubmit = this.handleSubmit.bind(this);
		this.onImportTransactions = this.handleImportTransactions.bind(this);
	}
	componentWillReceiveProps(nextProps) {
		if (nextProps.show && !this.props.show) {
			this.setState(this.getInitialState());
		}
	}
	handleCancel() {
		if (this.props.onCancel != null)
			this.props.onCancel();
	}
	handleImportChange() {
		this.setState({importFile: ReactDOM.findDOMNode(this.refs.importfile).value});
	}
	handleTypeChange(type) {
		this.setState({importType: type.TypeId});
	}
	handlePasswordChange() {
		this.setState({password: ReactDOM.findDOMNode(this.refs.password).value});
	}
	handleStartDateChange(date, string) {
		if (date == null)
			return;
		this.setState({
			startDate: date
		});
	}
	handleEndDateChange(date, string) {
		if (date == null)
			return;
		this.setState({
			endDate: date
		});
	}
	handleSubmit() {
		if (this.props.onSubmit != null)
			this.props.onSubmit(this.props.account);
	}
	handleImportTransactions() {
		if (this.state.importType == ImportType.OFX) {
			this.props.onImportOFX(this.props.account, this.state.password, this.state.startDate, this.state.endDate);
		} else if (this.state.importType == ImportType.OFXFile) {
			this.props.onImportOFXFile(ReactDOM.findDOMNode(this.refs.importfile), this.props.account);
		} else if (this.state.importType == ImportType.Gnucash) {
			this.props.onImportGnucash(ReactDOM.findDOMNode(this.refs.importfile));
		}
	}
	render() {
		var accountNameLabel = "Performing global import:"
		if (this.props.account != null && this.state.importType != ImportType.Gnucash)
			accountNameLabel = "Importing to '" + getAccountDisplayName(this.props.account, this.props.accounts) + "' account:";

		// Display the progress bar if an upload/import is in progress
		var progressBar = [];
		if (this.props.imports.importing && this.props.imports.uploadProgress == 100) {
			progressBar = (<ProgressBar now={this.props.imports.uploadProgress} active label="Importing transactions..." />);
		} else if (this.props.imports.importing) {
			progressBar = (<ProgressBar now={this.props.imports.uploadProgress} active label={`Uploading... ${this.props.imports.uploadProgress}%`} />);
		}

		// Create panel, possibly displaying error or success messages
		var panel = [];
		if (this.props.imports.importFailed) {
			panel = (<Panel header="Error Importing Transactions" bsStyle="danger">{this.props.imports.errorMessage}</Panel>);
		} else if (this.props.imports.importFinished) {
			panel = (<Panel header="Successfully Imported Transactions" bsStyle="success">Your import is now complete.</Panel>);
		}

		// Display proper buttons, possibly disabling them if an import is in progress
		var button1 = [];
		var button2 = [];
		if (!this.props.imports.importFinished && !this.props.imports.importFailed) {
			var importingDisabled = this.props.imports.importing || (this.state.importType != ImportType.OFX && this.state.importFile == "") || (this.state.importType == ImportType.OFX && this.state.password == "");
			button1 = (<Button onClick={this.onCancel} disabled={this.props.imports.importing} bsStyle="warning">Cancel</Button>);
			button2 = (<Button onClick={this.onImportTransactions} disabled={importingDisabled} bsStyle="success">Import</Button>);
		} else {
			button1 = (<Button onClick={this.onSubmit} disabled={this.props.imports.importing} bsStyle="success">OK</Button>);
		}
		var inputDisabled = (this.props.imports.importing || this.props.imports.importFailed || this.props.imports.importFinished) ? true : false;

		// Disable OFX/QFX imports if no account is selected
		var disabledTypes = false;
		if (this.props.account == null)
			disabledTypes = [ImportTypeList[ImportType.OFX - 1], ImportTypeList[ImportType.OFXFile - 1]];

		var importForm = [];
		if (this.state.importType == ImportType.OFX) {
			importForm = (
				<div>
				<FormGroup>
					<Col componentClass={ControlLabel} xs={2}>OFX Password</Col>
					<Col xs={10}>
						<FormControl type="password"
							value={this.state.password}
							placeholder="Password..."
							ref="password"
							onChange={this.onPasswordChange} />
					</Col>
				</FormGroup>
				<FormGroup>
					<Col componentClass={ControlLabel} xs={2}>Start Date</Col>
					<Col xs={10}>
					<DateTimePicker
						time={false}
						defaultValue={this.state.startDate}
						onChange={this.onStartDateChange} />
					</Col>
				</FormGroup>
				<FormGroup>
					<Col componentClass={ControlLabel} xs={2}>End Date</Col>
					<Col xs={10}>
					<DateTimePicker
						time={false}
						defaultValue={this.state.endDate}
						onChange={this.onEndDateChange} />
					</Col>
				</FormGroup>
				</div>
			);
		} else {
			importForm = (
				<FormGroup>
					<Col componentClass={ControlLabel} xs={2}>File</Col>
					<Col xs={10}>
					<FormControl type="file"
						ref="importfile"
						disabled={inputDisabled}
						value={this.state.importFile}
						onChange={this.onImportChange} />
					<HelpBlock>Select a file to upload.</HelpBlock>
					</Col>
				</FormGroup>
			);
		}

		return (
			<Modal show={this.props.show} onHide={this.onCancel}>
				<Modal.Header closeButton>
					<Modal.Title>Import Transactions</Modal.Title>
				</Modal.Header>
				<Modal.Body>
				<Form horizontal onSubmit={this.onImportTransactions}
						encType="multipart/form-data"
						ref="importform">
					<FormGroup>
					<Col xs={12}>
						<ControlLabel>{accountNameLabel}</ControlLabel>
					</Col>
					</FormGroup>
					<FormGroup>
					<Col componentClass={ControlLabel} xs={2}>Import Type</Col>
					<Col xs={10}>
					<DropdownList
						data={ImportTypeList}
						valueField='TypeId'
						textField='Name'
						onSelect={this.onTypeChange}
						defaultValue={this.state.importType}
						disabled={disabledTypes}
						ref="importtype" />
					</Col>
					</FormGroup>
					{importForm}
				</Form>
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
}

class AccountRegister extends React.Component {
	constructor() {
		super();
		this.state = {
			newTransaction: null,
			height: 0
		};
		this.onEditTransaction = this.handleEditTransaction.bind(this);
		this.onEditingCancel = this.handleEditingCancel.bind(this);
		this.onNewTransactionClicked = this.handleNewTransactionClicked.bind(this);
		this.onSelectPage = this.handleSelectPage.bind(this);
		this.onImportComplete = this.handleImportComplete.bind(this);
		this.onUpdateTransaction = this.handleUpdateTransaction.bind(this);
		this.onDeleteTransaction = this.handleDeleteTransaction.bind(this);
	}
	resize() {
		var div = ReactDOM.findDOMNode(this);
		this.setState({height: div.parentElement.clientHeight - 64});
	}
	componentWillReceiveProps(nextProps) {
		if (!nextProps.transactionPage.upToDate && nextProps.selectedAccount != -1) {
			nextProps.onFetchTransactionPage(nextProps.accounts[nextProps.selectedAccount], nextProps.transactionPage.pageSize, nextProps.transactionPage.page);
		}
	}
	componentDidMount() {
		this.resize();
		var self = this;
		$(window).resize(function() {self.resize();});
	}
	handleEditTransaction(transaction) {
		this.props.onSelectTransaction(transaction.TransactionId);
	}
	handleEditingCancel() {
		this.setState({
			newTransaction: null
		});
		this.props.onUnselectTransaction();
	}
	handleNewTransactionClicked() {
		var newTransaction = new Transaction();
		newTransaction.Date = new Date();
		newTransaction.Splits.push(new Split());
		newTransaction.Splits.push(new Split());
		newTransaction.Splits[0].Status = SplitStatus.Entered;
		newTransaction.Splits[1].Status = SplitStatus.Entered;
		newTransaction.Splits[0].AccountId = this.props.accounts[this.props.selectedAccount].AccountId;

		this.setState({
			newTransaction: newTransaction
		});
	}
	handleSelectPage(eventKey) {
		var newpage = eventKey - 1;
		// Don't do pages that don't make sense
		if (newpage < 0)
			newpage = 0;
		if (newpage >= this.props.transactionPage.numPages)
			newpage = this.props.transactionPage.numPages - 1;
		if (newpage != this.props.transactionPage.page) {
			if (this.props.selectedAccount != -1) {
				this.props.onFetchTransactionPage(this.props.accounts[this.props.selectedAccount], this.props.pageSize, newpage);
			}
		}
	}
	handleImportComplete() {
		this.props.onCloseImportModal();
		this.props.onFetchAllAccounts();
		this.props.onFetchAllSecurities();
		if (this.props.selectedAccount != -1) {
			this.props.onFetchTransactionPage(this.props.accounts[this.props.selectedAccount], this.props.pageSize, this.props.transactionPage.page);
		}
	}
	handleDeleteTransaction(transaction) {
		this.props.onDeleteTransaction(transaction);
		this.props.onUnselectTransaction();
	}
	handleUpdateTransaction(transaction) {
		if (transaction.TransactionId != -1) {
			this.props.onUpdateTransaction(transaction);
		} else {
			this.props.onCreateTransaction(transaction);
		}
		this.props.onUnselectTransaction();
		this.setState({
			newTransaction: null
		});
	}
	render() {
		var name = "Please select an account";
		var register = [];

		if (this.props.selectedAccount != -1) {
			name = this.props.accounts[this.props.selectedAccount].Name;

			var transactionRows = [];
			for (var i = 0; i < this.props.transactionPage.transactions.length; i++) {
				var transactionId = this.props.transactionPage.transactions[i];
				var transaction = this.props.transactions[transactionId];
				transactionRows.push((
					<TransactionRow
						key={transactionId}
						transaction={transaction}
						account={this.props.accounts[this.props.selectedAccount]}
						accounts={this.props.accounts}
						securities={this.props.securities}
						onSelect={this.props.onSelectTransaction}/>
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

		var disabled = (this.props.selectedAccount == -1) ? true : false;

		var transactionSelected = false;
		var selectedTransaction = new Transaction();
		if (this.state.newTransaction != null) {
			selectedTransaction = this.state.newTransaction;
			transactionSelected = true;
		} else if (this.props.transactionPage.selection != -1) {
			selectedTransaction = this.props.transactions[this.props.transactionPage.selection];
			transactionSelected = true;
		}

		return (
			<div className="transactions-container">
				<AddEditTransactionModal
					show={transactionSelected}
					transaction={selectedTransaction}
					accounts={this.props.accounts}
					accountChildren={this.props.accountChildren}
					onCancel={this.onEditingCancel}
					onSubmit={this.onUpdateTransaction}
					onDelete={this.onDeleteTransaction}
					securities={this.props.securities} />
				<ImportTransactionsModal
					imports={this.props.imports}
					show={this.props.imports.showModal}
					account={this.props.accounts[this.props.selectedAccount]}
					accounts={this.props.accounts}
					onCancel={this.props.onCloseImportModal}
					onHide={this.props.onCloseImportModal}
					onSubmit={this.onImportComplete}
					onImportOFX={this.props.onImportOFX}
					onImportOFXFile={this.props.onImportOFXFile}
					onImportGnucash={this.props.onImportGnucash} />
				<div className="transactions-register-toolbar">
				Transactions for '{name}'
				<ButtonToolbar className="pull-right">
					<ButtonGroup>
					<Pagination
						className="skinny-pagination"
						prev next first last ellipsis
						items={this.props.transactionPage.numPages}
						maxButtons={Math.min(5, this.props.transactionPage.numPages)}
						activePage={this.props.transactionPage.page + 1}
						onSelect={this.onSelectPage} />
					</ButtonGroup>
					<ButtonGroup>
					<Button
							onClick={this.onNewTransactionClicked}
							bsStyle="success"
							disabled={disabled}>
						<Glyphicon glyph='plus-sign' /> New Transaction
					</Button>
					<Button
							onClick={this.props.onOpenImportModal}
							bsStyle="primary">
						<Glyphicon glyph='import' /> Import
					</Button>
					</ButtonGroup>
				</ButtonToolbar>
				</div>
				{register}
			</div>
		);
	}
}

module.exports = AccountRegister;
