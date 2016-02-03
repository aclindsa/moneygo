#include <libofx/libofx.h>
#include "_cgo_export.h"

int ofx_statement_callback(const struct OfxStatementData statement_data, void *data) {
	return OFXStatementCallback(statement_data, data);
}

int ofx_account_callback(const struct OfxAccountData account_data, void *data) {
	return OFXAccountCallback(account_data, data);
}

int ofx_transaction_callback(const struct OfxTransactionData transaction_data, void *data) {
	return OFXTransactionCallback(transaction_data, data);
}
