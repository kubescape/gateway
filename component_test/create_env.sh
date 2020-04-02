#!/bin/bash
set -e

if ! python3 -m venv python_env; then
  echo "Failed to create python environment"
  exit 1
fi

. python_env/bin/activate
pip3 install docker requests bson websocket
deactivate
