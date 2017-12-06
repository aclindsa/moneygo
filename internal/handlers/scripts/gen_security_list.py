#!/usr/bin/env python

import csv
from xml.dom import minidom
import sys

if sys.version_info[0] < 3:
    from urllib2 import urlopen

    # Allow writing utf-8 to stdout
    import codecs
    UTF8Writer = codecs.getwriter('utf8')
    sys.stdout = UTF8Writer(sys.stdout)
else:
    from urllib.request import urlopen

    # This is absent, but also unneeded in python3, so just return the string
    def unicode(s, encoding):
        return s

class Security(object):
    def __init__(self, name, description, number, _type, precision):
        self.name = name
        self.description = description
        self.number = number
        self.type = _type
        self.precision = precision
    def unicode(self):
        s = """\t{
\t\tName: \"%s\",
\t\tDescription: \"%s\",
\t\tSymbol: \"%s\",
\t\tPrecision: %d,
\t\tType: %s,
\t\tAlternateId: \"%s\"},\n""" % (self.name, self.description, self.name, self.precision, self.type, str(self.number))
        try:
            return unicode(s, 'utf_8')
        except TypeError:
            return s

class SecurityList(object):
    def __init__(self, comment):
        self.comment = comment
        self.currencies = {}
    def add(self, currency):
        self.currencies[currency.number] = currency
    def unicode(self):
        string = "\t// "+self.comment+"\n"
        for key in sorted(self.currencies.keys()):
            string += self.currencies[key].unicode()
        return string

def process_ccyntry(currency_list, node):
    name = ""
    nameSet = False
    number = 0
    numberSet = False
    description = ""
    precision = 0
    for n in node.childNodes:
        if n.nodeName == "Ccy":
            name = n.firstChild.nodeValue
            nameSet = True
        elif n.nodeName == "CcyNm":
            description = n.firstChild.nodeValue
        elif n.nodeName == "CcyNbr":
            number = int(n.firstChild.nodeValue)
            numberSet = True
        elif n.nodeName == "CcyMnrUnts":
            if n.firstChild.nodeValue == "N.A.":
                precision = 0
            else:
                precision = int(n.firstChild.nodeValue)
    if nameSet and numberSet:
        currency_list.add(Security(name, description, number, "models.Currency", precision))

def get_currency_list():
    currency_list = SecurityList("ISO 4217, from http://www.currency-iso.org/en/home/tables/table-a1.html")

    f = urlopen('http://www.currency-iso.org/dam/downloads/lists/list_one.xml')
    xmldoc = minidom.parse(f)
    for isonode in xmldoc.childNodes:
        if isonode.nodeName == "ISO_4217":
            for ccytblnode in isonode.childNodes:
                if ccytblnode.nodeName == "CcyTbl":
                    for ccyntrynode in ccytblnode.childNodes:
                        if ccyntrynode.nodeName == "CcyNtry":
                            process_ccyntry(currency_list, ccyntrynode)
    f.close()
    return currency_list

def get_cusip_list(filename):
    cusip_list = SecurityList("")
    with open(filename) as csvfile:
        csvreader = csv.reader(csvfile, delimiter=',')
        for row in csvreader:
            cusip = row[0]
            name = row[1]
            description = ",".join(row[2:])
            cusip_list.add(Security(name, description, cusip, "models.Stock", 5))
    return cusip_list

def main():
    currency_list = get_currency_list()
    cusip_list = get_cusip_list('cusip_list.csv')

    print("package handlers\n")
    print("import (")
    print("\t\"github.com/aclindsa/moneygo/internal/models\"")
    print(")\n")
    print("var SecurityTemplates = []models.Security{")
    print(currency_list.unicode())
    print(cusip_list.unicode())
    print("}")

if __name__ == "__main__":
    main()
