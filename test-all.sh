#!/bin/bash

cd "/mnt/d/e drive/RentRelay"

echo "================================================"
echo "  RentRelay - Running all smoke tests"
echo "================================================"
echo ""

run_test() {
  echo -n "Testing $1... "
  result=$(go run -buildvcs=false ./cmd/$1-smoke 2>&1)
  if [ $? -eq 0 ]; then
    echo "PASSED ✅"
    echo "$result" | head -3
  else
    echo "FAILED ❌"
    echo "$result" | tail -3
  fi
  echo ""
}

run_test user
run_test property
run_test landlord
run_test tenant
run_test matching
run_test agreement
run_test notification
run_test document
run_test storage

echo "================================================"
echo "  All tests done!"
echo "================================================"
