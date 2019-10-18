#!/bin/bash
cd ~/git/$1 
go test ./... 2>&1 > $2.log
RESULT=$?
if [ $RESULT -eq 0 ]; then
  echo "PASS: $1. UUID: $2"
else
  echo "FAIL: $1. UUID: $2"
fi
exit 0