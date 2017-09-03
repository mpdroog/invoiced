#!/bin/bash
#set -x
set -e
set -u

rm -rf db
go build && ./hack
rm -rf ../../billingdb
#mv db/ ../../billingdb