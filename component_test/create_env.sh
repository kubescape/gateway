#!/bin/bash
set -e

if ! python3 -m venv test_env; then
  echo "Failed to create python environment"
  exit 1
fi

. test_env/bin/activate
pip3 install -r requirements.txt
deactivate
