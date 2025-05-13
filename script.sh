#!/bin/sh

echo -n "Program: "
for file in `ls *.go | grep -v _test`; do sed '/^[[:space:]]*$/d' $file | wc -l; done | awk '{n += $1}; END{print n}'

echo -n "Tests:   "
for file in `ls *_test.go`; do sed '/^[[:space:]]*$/d' $file | wc -l; done | awk '{n += $1}; END{print n}'

echo -n "Total:   "
cat *.go | sed '/^[[:space:]]*$/d' | wc -l | awk '{print $1}'
