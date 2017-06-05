#!/bin/bash
QUARTER=2017q1

function get_ticker() {
	local cusip=$1

	local tmpfile=$tmpdir/curl_tmpfile
	curl -s -d "sopt=cusip&tickersymbol=${cusip}" http://quantumonline.com/search.cfm > $tmpfile
	local quantum_name=$(sed -rn 's@<font size="\+1"><center><b>(.+)</b><br></center></font>\s*$@\1@p' $tmpfile | head -n1)
	local quantum_ticker=$(sed -rn 's@^.*Ticker Symbol: ([A-Z\.0-9\-]+) &nbsp;&nbsp;&nbsp;&nbsp;CUSIP.*$@\1@p' $tmpfile | head -n1)

	if [[ -z $quantum_ticker ]] || [[ -z $quantum_name ]]; then
		curl -s -d "reqforlookup=REQUESTFORLOOKUP&productid=mmnet&isLoggedIn=mmnet&rows=50&for=stock&by=cusip&criteria=${cusip}&submit=Search" http://quotes.fidelity.com/mmnet/SymLookup.phtml > $tmpfile
		fidelity_name=$(sed -rn 's@<tr><td height="20" nowrap><font class="smallfont">(.+)</font></td>\s*@\1@p' $tmpfile | sed -r 's/\&amp;/\&/')
		fidelity_ticker=$(sed -rn 's@\s+<td align="center" width="20%"><font><a href="/webxpress/get_quote\?QUOTE_TYPE=\&SID_VALUE_ID=(.+)">(.+)</a></td>\s*@\1@p' $tmpfile | head -n1)
		if [[ -z $fidelity_ticker ]] || [[ -z $fidelity_name ]]; then
			echo $cusip >> $tmpdir/${QUARTER}_bad_cusips.csv
		else
			echo "$cusip,$fidelity_ticker,$fidelity_name"
		fi
	else
		echo "$cusip,$quantum_ticker,$quantum_name"
	fi
}

tmpdir=$(mktemp -d -p $PWD)

# Get the list of CUSIPs from the SEC and generate a nicer format of it
wget -q http://www.sec.gov/divisions/investment/13f/13flist${QUARTER}.pdf -O $tmpdir/13flist${QUARTER}.pdf
pdftotext -layout $tmpdir/13flist${QUARTER}.pdf - > $tmpdir/13flist${QUARTER}.txt
sed -rn 's/^([A-Z0-9]{6}) ([A-Z0-9]{2}) ([A-Z0-9]) .*$/\1\2\3/p' $tmpdir/13flist${QUARTER}.txt > $tmpdir/${QUARTER}_cusips

# Find tickers and names for all the CUSIPs we can and print them out
for cusip in $(cat $tmpdir/${QUARTER}_cusips); do
	get_ticker $cusip
done

rm -rf $tmpdir
