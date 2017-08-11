InvoiceD
===========
Simple hour registration + invoice generator.

What makes us better than the competition:

* we save all userdata as TOML-files onto the filesystem (human readable)
* we commit changes to a local Git-repository (free revision control of precious data)
* we index everything (Bleve) so we got a power search!
* we're good looking
* we're opensource

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