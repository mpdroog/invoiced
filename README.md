InvoiceD
===========
Simple hour registration + invoice generator.

What makes us better than the competition (non-technical):

* Audit trail, every change is recorded and easily revertable!
* Always available, no internet? no problem!! The whole system runs from your own PC (if you want)
* ~~Searching for an old invoice? We got a powersearch, so finding it should be a breeze!~~
* We're FAST, instantly loading pages :D
* ~~Completely free! (But if you pay 1EUR/month we'll give you backups)~~
* Quick keys, you don't need to use your mouse all the time :)

What makes us better than the competition (technical):

* we save all userdata as TOML-files onto the filesystem (human readable and possibly readable after corruption)
* we commit changes to a local Git-repository (free version control of precious data with an audit-trail of changes)
* we push/pull with remote GIT nodes on change (distributed accounting/easy backups)
* ~~we index everything (Bleve) so we got a power search!~~
* we're good looking
* we're FAST (average page loading time of 400ms!)
* we're opensource (sourcecode is free, adjustable and we like contributions back)

Technical background
===========
- The backend is Golang/~~Bleve~~/Git
- The frontend is ReactJS/TypeScript/Brunch (static-src)

1] contrib/desktop( invoiced )
2] static

1] contrib/desktop
Implements the desktop specific stuff
and runs the invoice-daemon into a sub-process

2] static
Is offered by invoiced and contains all HTML+CSS+JS
that uses the invoiced API to build it's UI.

How to build
===========
Frontend
```
# Install NodeJS
# Install yarn (better npm)
cd static-src
yarn install
npm run dev
```

Backend
```
# Install go
go get github.com/mpdroog/invoiced
cd $GOPATH/src/github.com/mpdroog/invoiced
go build
```

Installing
```
# First create some keys
cd contrib/gen
$ go build && ./gen
Enter Password: <XXX>
entities.toml input
IV:   &d13dfa0685c88dd2df7a4afb4359ed8
Salt: 9c7cd736769046bd6779599e71785f65
Hash: 8476a5883163ada1ebfb80257a2303ec6167059a6329e212b561f37004009c5f
cd -

# Init db (~ is your homedir)
mkdir ~/billingdb
cd ~/billingdb
git init
vi entities.toml
>>
IV="PASTE_IV_HERE"
Version=1

[company.name]
Name="YOUR_COMPANY"
COC="YOUR_Chamber_Of_Commerce_Number"
VAT="YOUR_VAT_NUMBER"
IBAN="YOUR_IBAN_NUMBER"
BIC="YOUR_BANK_BIC"
Salt="PASTE_SALT_HERE"

[[user]]
Email="you@yourcompany.com"
Hash="PASTE_HASH_HERE"
Company=["name"]
Name="YourName"
Address1="Street houseno"
Address2="Postal City"
<<

mkdir -p name/$(date +\%Y) # mkdir name/2019
vi name/counters.toml
>>
InvoiceID = 1
<<
vi name/debtors.toml
>>
[debtorname]
Name="Debtorname B.V."
Street1="Street Houseno"
Street2="Postal City"
VAT=""
COC=""
TAX="NL21"
NoteAdd=""
BillingEmail=["finance@yourcustomer.nl"]
<<
vi name/projects.toml
>>
[projectname]
Debtor="debtorname"
BillingEmail=["accounting@importantcustomer.de"]
NoteAdd="Comment on invoice"
HourRate=20.00
DueDays=14
PO=""
Street1="AdditionalText"
<<
git add .
git commit -m "Initial commit"

cd -
vi config.toml
>>
[queues.support]
User = "myemail@gmail.com"
Pass = "supersecret"
Host = "smtp.gmail.com"
Port = 465
From = "myemail@gmail.com"
FromReply = "myemail@gmail.com"
Display = "MyCompany"
Subject = ""
BCC = ["myemail@gmail.com"]
>>

./invoiced -v -d=~/billingdb
open "http://localhost:9999"
```

