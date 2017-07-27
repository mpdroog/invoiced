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
brew install
npm run dll
npm run dev
```

Backend
```
# Install go
go get github.com/mpdroog/invoiced
cd $GOPATH/src/github.com/mpdroog/invoiced
go build
```

Run
```
./invoiced -v
open "http://localhost:9999/static"
```


