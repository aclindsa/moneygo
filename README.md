# MoneyGo

**MoneyGo** is a personal finance web application written in JavaScript and
Golang. It adheres to [double-entry
accounting](https://en.wikipedia.org/wiki/Double-entry_bookkeeping_system)
principles and allows for importing directly from financial institutions using
OFX (via [ofxgo](https://github.com/aclindsa/ofxgo)).

This project is in active development and is not yet ready to be relied upon as
your primary accounting software.

## Installation

First, install npm in your distribution:

	$ sudo pacman -S npm

Install browserify globally using npm:

	$ sudo npm install -g browserify

You'll then want to build everything (the Golang and Javascript portions) using
something like:

	$ export GOPATH=`pwd`
	$ go get -v github.com/aclindsa/moneygo
	$ go generate -v github.com/aclindsa/moneygo
	$ go install -v github.com/aclindsa/moneygo

This may take quite a while the first time you build the project since it is
auto-generating a list of currencies and securities by querying multiple
websites and services. To avoid this step, you can `touch
src/github.com/aclindsa/moneygo/cusip_list.csv` before executing the `go
generate ...` command above. Note that this will mean that no security templates
are available to easily populate securities in your installation. If you would
like to later generate these, simply remove the cusip_list.csv file and re-run
the `go generate ...` command.

## Running

Assuming you're in the same directory you ran the above installation commands
from, running MoneyGo is then as easy as:

	$ ./bin/moneygo \
	  -port 8080 \
	  -base src/github.com/aclindsa/moneygo/

You should then be able to explore MoneyGo by visiting http://localhost:8080 in
your browser.
