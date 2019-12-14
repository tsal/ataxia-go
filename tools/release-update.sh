#!/usr/bin/env bash

RELFILE=${RELFILE:-"cmd/ataxia/package.go"}

if [[ ! -d ".git" ]]; then
  echo "Must be run at the root of the git repository (.git directory not found)"
  exit 1
fi

DATE=`date`
RELEASE=${RELEASE:-$(git tag | egrep "^v([0-9]+\.?)+" | tail -n 1 | cut -d' ' -f 2)}

echo "Updating release constants in ${RELFILE}:"
echo "  ataxiaVersion  = '${RELEASE}'"
echo "  ataxiaCompiled = '${DATE}'"

cat >"${RELFILE}" <<EOF
package main

const (
	ataxiaVersion  = "${RELEASE}"
	ataxiaCompiled = "${DATE}"
)
EOF
