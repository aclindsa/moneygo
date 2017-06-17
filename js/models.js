var Big = require('big.js');

function getJSONObj(json_input) {
	if (typeof json_input == "string")
		return $.parseJSON(json_input)
	else if (typeof json_input == "object")
		return json_input;

	console.error("Unable to parse json:", json_input);
	return null
}

class User {
	constructor() {
		this.UserId = -1;
		this.Name = "";
		this.Username = "";
		this.Password = "";
		this.Email = "";
	}
	toJSON() {
		var json_obj = {};
		json_obj.UserId = this.UserId;
		json_obj.Name = this.Name;
		json_obj.Username = this.Username;
		json_obj.Password = this.Password;
		json_obj.Email = this.Email;
		return JSON.stringify(json_obj);
	}
	fromJSON(json_input) {
		var json_obj = getJSONObj(json_input);

		if (json_obj.hasOwnProperty("UserId"))
			this.UserId = json_obj.UserId;
		if (json_obj.hasOwnProperty("Name"))
			this.Name = json_obj.Name;
		if (json_obj.hasOwnProperty("Username"))
			this.Username = json_obj.Username;
		if (json_obj.hasOwnProperty("Password"))
			this.Password = json_obj.Password;
		if (json_obj.hasOwnProperty("Email"))
			this.Email = json_obj.Email;
	}
	isUser() {
		var empty_user = new User();
		return this.UserId != empty_user.UserId ||
			this.Username != empty_user.Username;
	}
}

class Session {
	constructor() {
		this.SessionId = -1;
		this.UserId = -1;
	}
	toJSON() {
		var json_obj = {};
		json_obj.SessionId = this.SessionId;
		json_obj.UserId = this.UserId;
		return JSON.stringify(json_obj);
	}
	fromJSON(json_input) {
		var json_obj = getJSONObj(json_input);

		if (json_obj.hasOwnProperty("SessionId"))
			this.SessionId = json_obj.SessionId;
		if (json_obj.hasOwnProperty("UserId"))
			this.UserId = json_obj.UserId;
	}
	isSession() {
		var empty_session = new Session();
		return this.SessionId != empty_session.SessionId ||
			this.UserId != empty_session.UserId;
	}
}

const SecurityType = {
	Currency: 1,
	Stock: 2
}
var SecurityTypeList = [];
for (var type in SecurityType) {
	if (SecurityType.hasOwnProperty(type)) {
		SecurityTypeList.push({'TypeId': SecurityType[type], 'Name': type});
   }
}

class Security {
	constructor() {
		this.SecurityId = -1;
		this.Name = "";
		this.Description = "";
		this.Symbol = "";
		this.Precision = -1;
		this.Type = -1;
		this.AlternateId = "";
	}
	toJSON() {
		var json_obj = {};
		json_obj.SecurityId = this.SecurityId;
		json_obj.Name = this.Name;
		json_obj.Description = this.Description;
		json_obj.Symbol = this.Symbol;
		json_obj.Precision = this.Precision;
		json_obj.Type = this.Type;
		json_obj.AlternateId = this.AlternateId;
		return JSON.stringify(json_obj);
	}
	fromJSON(json_input) {
		var json_obj = getJSONObj(json_input);

		if (json_obj.hasOwnProperty("SecurityId"))
			this.SecurityId = json_obj.SecurityId;
		if (json_obj.hasOwnProperty("Name"))
			this.Name = json_obj.Name;
		if (json_obj.hasOwnProperty("Description"))
			this.Description = json_obj.Description;
		if (json_obj.hasOwnProperty("Symbol"))
			this.Symbol = json_obj.Symbol;
		if (json_obj.hasOwnProperty("Precision"))
			this.Precision = json_obj.Precision;
		if (json_obj.hasOwnProperty("Type"))
			this.Type = json_obj.Type;
		if (json_obj.hasOwnProperty("AlternateId"))
			this.AlternateId = json_obj.AlternateId;
	}
	isSecurity() {
		var empty_account = new Security();
		return this.SecurityId != empty_account.SecurityId ||
			this.Type != empty_account.Type;
	}
}

const AccountType = {
	Bank: 1,
	Cash: 2,
	Asset: 3,
	Liability: 4,
	Investment: 5,
	Income: 6,
	Expense: 7,
	Trading: 8,
	Equity: 9,
	Receivable: 10,
	Payable: 11
}
var AccountTypeList = [];
for (var type in AccountType) {
	if (AccountType.hasOwnProperty(type)) {
		AccountTypeList.push({'TypeId': AccountType[type], 'Name': type});
   }
}

class Account {
	constructor() {
		this.AccountId = -1;
		this.UserId = -1;
		this.SecurityId = -1;
		this.ParentAccountId = -1;
		this.Type = -1;
		this.Name = "";

		this.OFXURL = "";
		this.OFXORG = "";
		this.OFXFID = "";
		this.OFXUser = "";
		this.OFXBankID = "";
		this.OFXAcctID = "";
		this.OFXAcctType = "";
		this.OFXClientUID = "";
		this.OFXAppID = "";
		this.OFXAppVer = "";
		this.OFXVersion = "";
		this.OFXNoIndent = false
	}
	toJSON() {
		var json_obj = {};
		json_obj.AccountId = this.AccountId;
		json_obj.UserId = this.UserId;
		json_obj.SecurityId = this.SecurityId;
		json_obj.ParentAccountId = this.ParentAccountId;
		json_obj.Type = this.Type;
		json_obj.Name = this.Name;
		json_obj.OFXURL = this.OFXURL;
		json_obj.OFXORG = this.OFXORG;
		json_obj.OFXFID = this.OFXFID;
		json_obj.OFXUser = this.OFXUser;
		json_obj.OFXBankID = this.OFXBankID;
		json_obj.OFXAcctID = this.OFXAcctID;
		json_obj.OFXAcctType = this.OFXAcctType;
		json_obj.OFXClientUID = this.OFXClientUID;
		json_obj.OFXAppID = this.OFXAppID;
		json_obj.OFXAppVer = this.OFXAppVer;
		json_obj.OFXVersion = this.OFXVersion;
		json_obj.OFXNoIndent = this.OFXNoIndent;
		return JSON.stringify(json_obj);
	}
	fromJSON(json_input) {
		var json_obj = getJSONObj(json_input);

		if (json_obj.hasOwnProperty("AccountId"))
			this.AccountId = json_obj.AccountId;
		if (json_obj.hasOwnProperty("UserId"))
			this.UserId = json_obj.UserId;
		if (json_obj.hasOwnProperty("SecurityId"))
			this.SecurityId = json_obj.SecurityId;
		if (json_obj.hasOwnProperty("ParentAccountId"))
			this.ParentAccountId = json_obj.ParentAccountId;
		if (json_obj.hasOwnProperty("Type"))
			this.Type = json_obj.Type;
		if (json_obj.hasOwnProperty("Name"))
			this.Name = json_obj.Name;
		if (json_obj.hasOwnProperty("OFXURL"))
			this.OFXURL = json_obj.OFXURL;
		if (json_obj.hasOwnProperty("OFXORG"))
			this.OFXORG = json_obj.OFXORG;
		if (json_obj.hasOwnProperty("OFXFID"))
			this.OFXFID = json_obj.OFXFID;
		if (json_obj.hasOwnProperty("OFXUser"))
			this.OFXUser = json_obj.OFXUser;
		if (json_obj.hasOwnProperty("OFXBankID"))
			this.OFXBankID = json_obj.OFXBankID;
		if (json_obj.hasOwnProperty("OFXAcctID"))
			this.OFXAcctID = json_obj.OFXAcctID;
		if (json_obj.hasOwnProperty("OFXAcctType"))
			this.OFXAcctType = json_obj.OFXAcctType;
		if (json_obj.hasOwnProperty("OFXClientUID"))
			this.OFXClientUID = json_obj.OFXClientUID;
		if (json_obj.hasOwnProperty("OFXAppID"))
			this.OFXAppID = json_obj.OFXAppID;
		if (json_obj.hasOwnProperty("OFXAppVer"))
			this.OFXAppVer = json_obj.OFXAppVer;
		if (json_obj.hasOwnProperty("OFXVersion"))
			this.OFXVersion = json_obj.OFXVersion;
		if (json_obj.hasOwnProperty("OFXNoIndent"))
			this.OFXNoIndent = json_obj.OFXNoIndent;
	}
	isAccount() {
		var empty_account = new Account();
		return this.AccountId != empty_account.AccountId ||
			this.UserId != empty_account.UserId;
	}
	isRootAccount() {
		var empty_account = new Account();
		return this.ParentAccountId == empty_account.ParentAccountId;
	}
}

const SplitStatus = {
	Imported: 1,
	Entered: 2,
	Cleared: 3,
	Reconciled: 4,
	Voided: 5
}
var SplitStatusList = [];
for (var type in SplitStatus) {
	if (SplitStatus.hasOwnProperty(type)) {
		SplitStatusList.push({'StatusId': SplitStatus[type], 'Name': type});
   }
}
var SplitStatusMap = {};
for (var status in SplitStatus) {
	if (SplitStatus.hasOwnProperty(status)) {
		SplitStatusMap[SplitStatus[status]] = status;
   }
}

class Split {
	constructor() {
		this.SplitId = -1;
		this.TransactionId = -1;
		this.AccountId = -1;
		this.SecurityId = -1;
		this.Status = -1;
		this.Number = "";
		this.Memo = "";
		this.Amount = new Big(0.0);
		this.Debit = false;
	}
	toJSONobj() {
		var json_obj = {};
		json_obj.SplitId = this.SplitId;
		json_obj.TransactionId = this.TransactionId;
		json_obj.AccountId = this.AccountId;
		json_obj.SecurityId = this.SecurityId;
		json_obj.Status = this.Status;
		json_obj.Number = this.Number;
		json_obj.Memo = this.Memo;
		json_obj.Amount = this.Amount.toFixed();
		json_obj.Debit = this.Debit;
		return json_obj;
	}
	fromJSONobj(json_obj) {
		if (json_obj.hasOwnProperty("SplitId"))
			this.SplitId = json_obj.SplitId;
		if (json_obj.hasOwnProperty("TransactionId"))
			this.TransactionId = json_obj.TransactionId;
		if (json_obj.hasOwnProperty("AccountId"))
			this.AccountId = json_obj.AccountId;
		if (json_obj.hasOwnProperty("SecurityId"))
			this.SecurityId = json_obj.SecurityId;
		if (json_obj.hasOwnProperty("Status"))
			this.Status = json_obj.Status;
		if (json_obj.hasOwnProperty("Number"))
			this.Number = json_obj.Number;
		if (json_obj.hasOwnProperty("Memo"))
			this.Memo = json_obj.Memo;
		if (json_obj.hasOwnProperty("Amount"))
			this.Amount = new Big(json_obj.Amount);
		if (json_obj.hasOwnProperty("Debit"))
			this.Debit = json_obj.Debit;
	}
	isSplit() {
		var empty_split = new Split();
		return this.SplitId != empty_split.SplitId ||
			this.TransactionId != empty_split.TransactionId ||
			this.AccountId != empty_split.AccountId ||
			this.SecurityId != empty_split.SecurityId;
	}
}

class Transaction {
	constructor() {
		this.TransactionId = -1;
		this.UserId = -1;
		this.Description = "";
		this.Date = new Date();
		this.Splits = [];
	}
	toJSON() {
		var json_obj = {};
		json_obj.TransactionId = this.TransactionId;
		json_obj.UserId = this.UserId;
		json_obj.Description = this.Description;
		json_obj.Date = this.Date.toJSON();
		json_obj.Splits = [];
		for (var i = 0; i < this.Splits.length; i++)
			json_obj.Splits.push(this.Splits[i].toJSONobj());
		return JSON.stringify(json_obj);
	}
	fromJSON(json_input) {
		var json_obj = getJSONObj(json_input);

		if (json_obj.hasOwnProperty("TransactionId"))
			this.TransactionId = json_obj.TransactionId;
		if (json_obj.hasOwnProperty("UserId"))
			this.UserId = json_obj.UserId;
		if (json_obj.hasOwnProperty("Description"))
			this.Description = json_obj.Description;
		if (json_obj.hasOwnProperty("Date")) {
			this.Date = json_obj.Date
			if (typeof this.Date === 'string') {
				var t = Date.parse(this.Date);
				if (t)
					this.Date = new Date(t);
				else
					this.Date = new Date(0);
			} else
				this.Date = new Date(0);
		}
		if (json_obj.hasOwnProperty("Splits")) {
			for (var i = 0; i < json_obj.Splits.length; i++) {
				var s = new Split();
				s.fromJSONobj(json_obj.Splits[i]);
				this.Splits.push(s);
			}
		}
	}
	isTransaction() {
		var empty_transaction = new Transaction();
		return this.TransactionId != empty_transaction.TransactionId ||
			this.UserId != empty_transaction.UserId;
	}
	deepCopy() {
		var t = new Transaction();
		t.fromJSON(this.toJSON());
		return t;
	}
	imbalancedSplitSecurities(account_map) {
		// Return a list of SecurityIDs for those securities that aren't balanced
		// in this transaction's splits. If a split's AccountId is invalid, that
		// split is ignored, so those must be checked elsewhere
		var splitBalances = {};
		const emptySplit = new Split();
		for (var i = 0; i < this.Splits.length; i++) {
			var split = this.Splits[i];
			var securityId = -1;
			if (split.AccountId != emptySplit.AccountId) {
				securityId = account_map[split.AccountId].SecurityId;
			} else if (split.SecurityId != emptySplit.SecurityId) {
				securityId = split.SecurityId;
			} else {
				continue;
			}
			if (securityId in splitBalances) {
				splitBalances[securityId] = split.Amount.plus(splitBalances[securityId]);
			} else {
				splitBalances[securityId] = split.Amount.plus(0);
			}
		}
		var imbalancedIDs = [];
		for (var id in splitBalances) {
			if (!splitBalances[id].eq(0)) {
				imbalancedIDs.push(id);
			}
		}
		return imbalancedIDs;
	}
}

class Error {
	constructor() {
		this.ErrorId = -1;
		this.ErrorString = "";
	}
	toJSON() {
		var json_obj = {};
		json_obj.ErrorId = this.ErrorId;
		json_obj.ErrorString = this.ErrorString;
		return JSON.stringify(json_obj);
	}
	fromJSON(json_input) {
		var json_obj = getJSONObj(json_input);

		if (json_obj.hasOwnProperty("ErrorId"))
			this.ErrorId = json_obj.ErrorId;
		if (json_obj.hasOwnProperty("ErrorString"))
			this.ErrorString = json_obj.ErrorString;
	}
	isError() {
		var empty_error = new Error();
		return this.ErrorId != empty_error.ErrorId ||
			this.ErrorString != empty_error.ErrorString;
	}
}

class Series {
	constructor() {
		this.Values = [];
		this.Series = {};
	}
	toJSONobj() {
		var json_obj = {};
		json_obj.Values = this.Values;
		json_obj.Series = {};
		for (var child in this.Series) {
			if (this.Series.hasOwnProperty(child))
				json_obj.Series[child] = this.Series[child].toJSONobj();
		}
		return json_obj;
	}
	fromJSONobj(json_obj) {
		if (json_obj.hasOwnProperty("Values"))
			this.Values = json_obj.Values;
		if (json_obj.hasOwnProperty("Series")) {
			for (var child in json_obj.Series) {
				if (json_obj.Series.hasOwnProperty(child))
					this.Series[child] = new Series();
					this.Series[child].fromJSONobj(json_obj.Series[child]);
			}
		}
	}
	mapReduceChildren(mapFn, reduceFn) {
		var children = {}
		for (var child in this.Series) {
			if (this.Series.hasOwnProperty(child))
				children[child] = this.Series[child].mapReduce(mapFn, reduceFn);
		}
		return children;
	}
	mapReduce(mapFn, reduceFn) {
		var childValues = [];
		if (mapFn)
			childValues.push(this.Values.map(mapFn));
		else
			childValues.push(this.Values.slice());

		for (var child in this.Series) {
			if (this.Series.hasOwnProperty(child))
				childValues.push(this.Series[child].mapReduce(mapFn, reduceFn));
		}

		var reducedValues = [];
		if (reduceFn && childValues.length > 0 && childValues[0].length > 0) {
			for (var j = 0; j < childValues[0].length; j++) {
				reducedValues.push(childValues.reduce(function(accum, curr, i, arr) {
					return reduceFn(accum, arr[i][j]);
				}, 0));
			}
		}

		return reducedValues;
	}
}

class Tabulation {
	constructor() {
		this.ReportId = "";
		this.Title = "";
		this.Subtitle = "";
		this.Units = "";
		this.Labels = [];
		this.Series = {};
		this.FlattenedSeries = {};
	}
	static topLevelSeriesName() {
		return "(top level)"
	}
	toJSON() {
		var json_obj = {};
		json_obj.ReportId = this.ReportId;
		json_obj.Title = this.Title;
		json_obj.Subtitle = this.Subtitle;
		json_obj.Units = this.Units;
		json_obj.Labels = this.Labels;
		json_obj.Series = {};
		for (var series in this.Series) {
			if (this.Series.hasOwnProperty(series))
				json_obj.Series[series] = this.Series[series].toJSONobj();
		}
		return JSON.stringify(json_obj);
	}
	fromJSON(json_input) {
		var json_obj = getJSONObj(json_input)

		if (json_obj.hasOwnProperty("ReportId"))
			this.ReportId = json_obj.ReportId;
		if (json_obj.hasOwnProperty("Title"))
			this.Title = json_obj.Title;
		if (json_obj.hasOwnProperty("Subtitle"))
			this.Subtitle = json_obj.Subtitle;
		if (json_obj.hasOwnProperty("Units"))
			this.Units = json_obj.Units;
		if (json_obj.hasOwnProperty("Labels"))
			this.Labels = json_obj.Labels;
		if (json_obj.hasOwnProperty("Series")) {
			for (var series in json_obj.Series) {
				if (json_obj.Series.hasOwnProperty(series))
					this.Series[series] = new Series();
					this.Series[series].fromJSONobj(json_obj.Series[series]);
			}
		}
	}
	mapReduceChildren(mapFn, reduceFn) {
		var series = {}
		for (var child in this.Series) {
			if (this.Series.hasOwnProperty(child))
				series[child] = this.Series[child].mapReduce(mapFn, reduceFn);
		}
		return series;
	}
	mapReduceSeries(mapFn, reduceFn) {
		return this.mapReduceChildren(mapFn, reduceFn);
	}
}

class OFXDownload {
	constructor() {
		this.OFXPassword = "";
		this.StartDate = new Date();
		this.EndDate = new Date();
	}
	toJSON() {
		var json_obj = {};
		json_obj.OFXPassword = this.OFXPassword;
		json_obj.StartDate = this.StartDate.toJSON();
		json_obj.EndDate = this.EndDate.toJSON();
		return JSON.stringify(json_obj);
	}
}

module.exports = {
	// Classes
	User: User,
	Session: Session,
	Security: Security,
	Account: Account,
	Split: Split,
	Transaction: Transaction,
	Tabulation: Tabulation,
	OFXDownload: OFXDownload,
	Error: Error,

	// Enums, Lists
	AccountType: AccountType,
	AccountTypeList: AccountTypeList,
	SecurityType: SecurityType,
	SecurityTypeList: SecurityTypeList,
	SplitStatus: SplitStatus,
	SplitStatusList: SplitStatusList,
	SplitStatusMap: SplitStatusMap,

	// Constants
	BogusPassword: "password"
};
