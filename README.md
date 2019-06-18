# League Accounts

A Webapp to store and share League of Legends accounts.

# Install

After installing Go and configuring $GOPATH, run:

```
go get github.com/erikfastermann/league-accounts
```

# Docker

After installing Docker and Docker-compose, run:

```
sudo docker-compose up -d
```

# Add a User

Add a user with the manage-users.go script by running:

```
cd $GOPATH/src/github.com/erikfastermann/league-accounts/scripts
go get
go run manage-users.go -a $LEAGUE_ACCS_DB_PATH
```

If you don't supply a path, the default for this docker volume will be used.

You might need sudo for this to work.

# HTTPS (Production)

To activate HTTPS, comment out the following in the docker-compose.yml file:

```yaml
LEAGUE_ACCS_PROD: 'true'
LEAGUE_ACCS_PROD_DOMAINS: your-domain.com,www.your-domain.com
```

# Appendix

## List of environment variables

Port (e.g.: 8080): LEAGUE_ACCS_PORT

Template Glob (e.g.: template/*): LEAGUE_ACCS_TEMPLATE_GLOB

Sqlite3 DB Path (e.g.: /league.db): LEAGUE_ACCS_DB_PATH

### HTTPS (Production)

Production (any value = true): LEAGUE_ACCS_PROD

Domains (e.g.: example.com,www.example.com): LEAGUE_ACCS_PROD_DOMAINS
