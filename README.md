# League Accounts

A Webapp to store and share League of Legends accounts.

# Getting started

## Docker

Clone the repository, install Docker and Docker-compose, then run:

```
sudo docker-compose up -d
```

**IMPORTANT**: Change the default username and passwords in the docker.env file.

### HTTPS (Production)

To activate HTTPS, add the following in the docker-compose.yml file (webapp -> environment):

```yaml
LEAGUE_ACCS_PROD: "true"
LEAGUE_ACCS_PROD_DOMAINS: your-domain.com,www.your-domain.com
```

## Manual

After installing Go and configuring GOPATH, run:

```
go get github.com/erikfastermann/league-accounts
```

### Database (MySQL)

After installing a MySQL Server (e.g.: MariaDB), login as the root user and run the following:

```sql
CREATE USER 'DB_USER'@'SERVER' IDENTIFIED BY 'SECURE_PASSWORD';
CREATE DATABASE DB_NAME;
USE DB_NAME;
GRANT ALL PRIVILEGES ON DB_NAME.* TO 'DB_USER'@'SERVER';
```

### Create a user

Generate a salted password hash.

```
go run $GOPATH/src/github.com/erikfastermann/league-accounts/scripts/gen-pass.go PASSWORD
```

Copy the output to your clipboard. Then login to your MySQL Server and run the following:

```sql
USE DB_NAME;
INSERT INTO users (Username, Password, Token) VALUES
('WEB_USERNAME', 'COPIED_PASSWORD_HASH', '');
```

### Testing

Export the environment variables listed below, then run:

```
go run $GOPATH/src/github.com/erikfastermann/league-accounts
```

# List of environment variables

Port (e.g.: 8080): LEAGUE_ACCS_PORT

Template Dir (e.g.: template/*): LEAGUE_ACCS_TEMPLATE_DIR

DB User: MYSQL_USER

DB Password: MYSQL_PASSWORD

DB Address (e.g.: localhost:3306): LEAGUE_ACCS_DB_ADDRESS

DB Name (e.g.: lol_accs): MYSQL_DATABASE

## HTTPS (Production)

Production (any value = true): LEAGUE_ACCS_PROD

Domains (e.g.: example.com,www.example.com): LEAGUE_ACCS_PROD_DOMAINS
