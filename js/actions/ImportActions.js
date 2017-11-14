var ImportConstants = require('../constants/ImportConstants');

var models = require('../models.js');
var OFXDownload = models.OFXDownload;
var Error = models.Error;

function beginImport() {
	return {
		type: ImportConstants.BEGIN_IMPORT
	}
}

function updateProgress(progress) {
	return {
		type: ImportConstants.UPDATE_IMPORT_PROGRESS,
		progress: progress
	}
}

function importFinished() {
	return {
		type: ImportConstants.IMPORT_FINISHED
	}
}

function importFailed(error) {
	return {
		type: ImportConstants.IMPORT_FAILED,
		error: error
	}
}

function openModal() {
	return function(dispatch) {
		dispatch({
			type: ImportConstants.OPEN_IMPORT_MODAL
		});
	};
}

function closeModal() {
	return function(dispatch) {
		dispatch({
			type: ImportConstants.CLOSE_IMPORT_MODAL
		});
	};
}

function importOFX(account, password, startDate, endDate) {
	return function(dispatch) {
		dispatch(beginImport());
		dispatch(updateProgress(50));

		var ofxdownload = new OFXDownload();
		ofxdownload.OFXPassword = password;
		ofxdownload.StartDate = startDate;
		ofxdownload.EndDate = endDate;

		$.ajax({
			type: "POST",
			contentType: "application/json",
			dataType: "json",
			url: "v1/accounts/"+account.AccountId+"/imports/ofx",
			data: ofxdownload.toJSON(),
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					var errString = e.ErrorString;
					if (e.ErrorId == 3 /* Invalid Request */) {
						errString = "Please check that your password and all other OFX login credentials are correct.";
					}
					dispatch(importFailed(errString));
				} else {
					dispatch(importFinished());
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(importFailed(error));
			}
		});
	};
}

function importFile(url, inputElement) {
	return function(dispatch) {
		dispatch(beginImport());

		if (inputElement.files.length == 0) {
			dispatch(importFailed("No files specified to be imported"))
			return;
		}
		if (inputElement.files.length > 1) {
			dispatch(importFailed("More than one file specified for import, only one allowed at a time"))
			return;
		}

		var file = inputElement.files[0];
		var formData = new FormData();
		formData.append('importfile', file, file.name);

		var handleSetProgress = function(e) {
			if (e.lengthComputable) {
				var pct = Math.round(e.loaded/e.total*100);
				dispatch(updateProgress(pct));
			} else {
				dispatch(updateProgress(50));
			}
		}

		$.ajax({
			type: "POST",
			url: url,
			data: formData,
			xhr: function() {
				var xhrObject = $.ajaxSettings.xhr();
				if (xhrObject.upload) {
					xhrObject.upload.addEventListener('progress', handleSetProgress, false);
				} else {
					dispatch(importFailed("File upload failed because xhr.upload isn't supported by your browser."));
				}
				return xhrObject;
			},
			success: function(data, status, jqXHR) {
				var e = new Error();
				e.fromJSON(data);
				if (e.isError()) {
					var errString = e.ErrorString;
					if (e.ErrorId == 3 /* Invalid Request */) {
						errString = "Please check that the file you uploaded is valid and try again.";
					}
					dispatch(importFailed(errString));
				} else {
					dispatch(importFinished());
				}
			},
			error: function(jqXHR, status, error) {
				dispatch(importFailed(error));
			},
			// So jQuery doesn't try to process the data or content-type
			cache: false,
			contentType: false,
			processData: false
		});
	};
}

function importOFXFile(inputElement, account) {
	var url = "v1/accounts/"+account.AccountId+"/imports/ofxfile";
	return importFile(url, inputElement);
}

function importGnucash(inputElement) {
	var url = "v1/imports/gnucash";
	return importFile(url, inputElement);
}

module.exports = {
	openModal: openModal,
	closeModal: closeModal,
	importOFX: importOFX,
	importOFXFile: importOFXFile,
	importGnucash: importGnucash
};
