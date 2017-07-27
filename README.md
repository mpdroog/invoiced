InvoiceD
===========
Simple hour registration + invoice generator.

Technical background
===========
- The backend is written in Golang, the used DB is BoltDB
- The frontend is ReactJS/TypeScript/Webpack

How to build
===========
Frontend
```
# Install NodeJS
cd static-src
npm install
npm install-dev
npm run dev
```

Backend
```
# Install go
go get github.com/mpdroog/invoiced
cd $GOPATH/src/github.com/mpdroog/invoiced
go build
# Webserver runnable with ./invoiced -v
```

Run
```
./invoiced -v
open "http://localhost:9999/static"
```


