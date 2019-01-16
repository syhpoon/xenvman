#!/bin/bash

set -e

CRDB="/cockroach/cockroach"

if [ "${1-}" = "shell" ]; then
  shift
  exec /bin/sh "$@"
fi

if [ -f /init.sql ]; then
   ${CRDB} start --insecure &

   until ${CRDB} sql --insecure < /init.sql; do sleep 1; done

   echo "*** DB populated"

   wait ${!}
else
  exec ${CRDB} start --insecure
fi
