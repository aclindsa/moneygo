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

## OFX Connection Details

The first (and potentially most tedious) step is to enter the OFX connection
details for your FI into the 'Sync (OFX)' tab when editing the MoneyGo account
for which you wish to import transactions.

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

**Account Type**: The type of account.

### Advanced Settings

TODO
