#!/bin/bash
cd ~/git/$1
go build $2 2>&1 
RESULT=$?
if [ $RESULT -eq 0 ]; then
  echo "PASS: $2"
else
  echo "FAIL: $2"
fi
exit 0