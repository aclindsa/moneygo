# OFX Imports

In the US, OFX ([official website](http://ofx.net/),
[Wikipedia](https://en.wikipedia.org/wiki/Open_Financial_Exchange)) is the most
widely-supported mechanism made available to customers of financial institutions
(FIs) to quickly import transactions and account balances. There are two ways to
import transactions (your FI may support one, both, or neither): one where the
customer downloads a .ofx file from the FI and then imports it into their
accounting software of choice, and one where the user provides accounting
software with their login credentials and the software negotiates the download
on their behalf. MoneyGo supports both import methods with the help of the
[ofxgo](https://github.com/aclindsa/ofxgo) project.

To import an OFX file you have already downloaded from your FI, click the
'Import' button after selecting the account you want to import transactions to,
select 'OFX/QFX File' as the Import Type, and upload the OFX file.

To import OFX transactions directly from the FI, instead select 'OFX' as the
Import Type and enter your password and the date range of transactions you want
to import. Note that there are a number of connections details that you will
need to fill out before this will work (see below).

## OFX Connection Details

The most tedious part is entering the OFX connection details for your FI into
the 'Sync (OFX)' tab when editing the MoneyGo account for which you wish to
import transactions.

### Helpful Sources

In my experience, FI's are not traditionally helpful when customers attempt to
connect to their OFX servers using accounting software other than the single
most popular one (though I encourage you to contact them to solicit this
information if for no other reason than to remind them that there is more than
one developer of consumer accounting software). Instead, we have to rely on
community efforts to piece together the connection details. Here are the sources
I've found most helpful:

* [https://wiki.gnucash.org/wiki/OFX_Direct_Connect_Bank_Settings](https://wiki.gnucash.org/wiki/OFX_Direct_Connect_Bank_Settings)
* [http://ofxhome.com/](http://ofxhome.com/)
* [https://ofxblog.wordpress.com/](https://ofxblog.wordpress.com/) (this
  contains the most out-of-date information, but may possibly be helpful if you
  can't find your FI at the other two sites)

### Fields

The following are the fields that make up the OFX connection details that
MoneyGo needs to know in order to successfully import transactions from your FI.

**OFX URL**: This is the URL that MoneyGo/ofxgo should make their initial
requests against. This is called 'Server URL' in the Gnucash wiki above, and FI
Url at ofxhome.

**ORG** and **FI**: These are sometimes seemingly-meaningless strings or numbers
that identify your particular financial institution from others which may share
the same URL. They are called 'FI Org' and 'FI Id' at ofxhome.

**Username**: This is your username, usually the same username you use to login
to your bank's website.

**Bank ID**: This is another identifier, specific to bank accounts (i.e. not
credit cards or investment accounts). It is frequently empty, or sometimes is
your bank's routing number or some other string.

**Account ID**: This is specific to your account, and is frequently your account
number.

**Account Type**: The type of account (i.e. 'Checking' or 'Savings')

### Advanced Settings

The following fields may or may not be required. Some of them are are to work
around incompatibilities with FI's OFX implementations, set the OFX version
spoken to something other than the default, or provide additional required
information.

**Client UID**: This is required by some banks (Chase) as a sort of two-factor
authentication. You may have to generate this yourself (using `uuidgen` on Linux
machines), and then some type of out-of-band approval/authentication to your FI
after you try to download transactions using OFX in MoneyGo for the first time. 

**App ID**: This is supposed to identify the application making the OFX
requests. ofxgo defaults to its own string if this is empty, but some FI's
require this to be a particular string in order to return valid results
(frequently 'QWIN')

**App Version**: The application version supposedly being used make the OFX
requests. Like they do for the 'App ID', you may have to set this to a
particular value to make your FI's server happy.

**OFX Version**: This defaults to "203" (for version 2.0.3 of the OFX spec) if
left empty. It must be one of "102", "103", "151", "160", "200", "201", "202",
"203", "210", "211", "220" if specified. This controls what version of OFX is
used to talk with the server.

**Don't indent OFX request files**: This is unchecked by default. Though rare,
some FI's implementations break if the SGML/XML elements are indented (and
others' break if they aren't!).
