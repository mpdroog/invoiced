InvoiceD
===========
Simple hour registration + invoice generator.

What makes us better than the competition (non-technical):

* Audit trail, every change is recorded and easily revertable!
* Always available, no internet? no problem!! The whole system runs from your own PC (if you want)
* Searching for an old invoice? We got a powersearch, so finding it should be a breeze!
* We're FAST, instantly loading pages :D
* Completely free! (But if you pay 1EUR/month we'll give you backups)

What makes us better than the competition (technical):

* we save all userdata as TOML-files onto the filesystem (human readable and possibly readable after corruption)
* we commit changes to a local Git-repository (free version control of precious data with an audit-trail of changes)
* we push/pull with remote GIT nodes on change (distributed accounting/easy backups)
* we index everything (Bleve) so we got a power search!
* we're good looking
* we're FAST (average page loading time of 400ms!)
* we're opensource (sourcecode is free, adjustable and we like contributions back)

Technical background
===========
- The backend is Golang/Bleve/Git
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

Run (dev-mode)
```
./invoiced -v
open "http://localhost:9999/static"
```

Run (desktop-mode)
```
cd contrib/desktop
go build
./desktop
```