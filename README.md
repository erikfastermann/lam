# League Accounts

A Webapp to store and share League of Legends accounts.

# Getting started

## Docker

Clone the repository, install Docker and Docker-compose, then run:

```
sudo docker-compose up -d
```

### HTTPS (Production)

To activate HTTPS, comment out the following in the docker-compose.yml file:

```yaml
LEAGUE_ACCS_PROD: 'true'
LEAGUE_ACCS_PROD_DOMAINS: your-domain.com,www.your-domain.com
```

## Manual

After installing Go and configuring GOPATH, run:

```
go get github.com/erikfastermann/league-accounts
```

### Create a user

Generate a salted password hash.

```
go run $GOPATH/src/github.com/erikfastermann/league-accounts/scripts/gen-pass.go PASSWORD
```

Copy the output to your clipboard. Then start the sqlite3 console and run the following:

```sql
INSERT INTO users (username, password, token) VALUES
('WEB_USERNAME', 'COPIED_PASSWORD_HASH', '');
```

### Testing

Export the environment variables listed below, then run:

```
go run $GOPATH/src/github.com/erikfastermann/league-accounts
```

# List of environment variables

Port (e.g.: 8080): LEAGUE_ACCS_PORT

Template Glob (e.g.: template/*): LEAGUE_ACCS_TEMPLATE_GLOB

DB Path (e.g.: /league.db): LEAGUE_ACCS_DB_PATH

## HTTPS (Production)

Production (any value = true): LEAGUE_ACCS_PROD

Domains (e.g.: example.com,www.example.com): LEAGUE_ACCS_PROD_DOMAINS
