# LAM - LoL Account Manager

A Webapp to store and share League of Legends accounts.

# Install

After installing Go and configuring $GOPATH, run:

```
go get github.com/erikfastermann/lam
```

# Docker

After installing docker and docker-compose, run:

```
sudo docker-compose up -d
```

# List of environment variables

Address (used for redirecting, e.g.: ':80'): `LAM_ADDRESS`

HTTPS Address: `LAM_HTTPS_ADDRESS`

Domain (used to upgrade from HTTP to HTTPS): `LAM_DOMAIN`

Cert file: `LAM_CERT`

Key file: `LAM_KEY`

Users (e.g.: 'user1:bcrypt1:user2:bcrypt2'): `LAM_USERS`

CSV DB Accounts (e.g.: '/accounts.csv'): `LAM_ACCOUNTS`

Template Glob (e.g.: 'template/*'): `LAM_TEMPLATE_GLOB`
