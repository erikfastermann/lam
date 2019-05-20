# League Accounts

A Webapp to store and share League of Legends accounts.

# Getting started

## Download

```
go get github.com/erikfastermann/league-accounts
```

## Database (MySQL)

After installing a MySQL Server (e.g.: MariaDB), login as the root user and run the following:

```sql
CREATE USER 'DB_USER'@'SERVER' IDENTIFIED BY 'SECURE_PASSWORD';
CREATE DATABASE DB_NAME;
USE DB_NAME;
GRANT ALL PRIVILEGES ON DB_NAME.* TO 'DB_USER'@'SERVER';
```

Then run `source scripts/db.sql;` to create the tables.

### Create a user

Generate a salted password hash.

```
go run $GOPATH/src/github.com/erikfastermann/league-accounts/scripts/gen-pass.go PASSWORD
```

Copy the output to your clipboard.

Then login to your MySQL Server and run the following:

```sql
USE DB_NAME;
INSERT INTO users (Username, Password, Token) VALUES
('WEB_USERNAME', 'COPIED_PASSWORD_HASH', '');
```

## Testing

```
$GOPATH/src/github.com/erikfastermann/league-accounts/scripts/run.sh
```

You might have to change some of the environment variables to fit your setup.

# List of environment variables

Template Dir (e.g.: template/*): LEAGUE_ACCS_TEMPLATE_DIR

DB User: LEAGUE_ACCS_DB_USER

DB Password: LEAGUE_ACCS_DB_PASSWORD

DB Address (e.g.: localhost:3306): LEAGUE_ACCS_DB_ADDRESS

DB Name (e.g.: lol_accs): LEAGUE_ACCS_DB_NAME

