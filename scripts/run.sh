#!/bin/bash

# Testing the server with some default variables

set -e

export LEAGUE_ACCS_DB_USER='erik'
export LEAGUE_ACCS_DB_PASSWORD='testpass'
export LEAGUE_ACCS_DB_ADDRESS='localhost:3306'
export LEAGUE_ACCS_DB_NAME="lol_accs"
export LEAGUE_ACCS_TEMPLATE_DIR="${GOPATH}/src/github.com/erikfastermann/league-accs/template/*"

go install github.com/erikfastermann/league-accs
league-accounts
