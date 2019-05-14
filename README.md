# League Accs

## Environment vars

Template Dir (e.g.: /tmp/*): LEAGUE_ACCS_TEMPLATE_DIR
DB User: LEAGUE_ACCS_DB_USER
DB Password: LEAGUE_ACCS_DB_PASSWORD
DB Address (e.g.: localhost:3306): LEAGUE_ACCS_DB_ADDRESS

## Database (mysql)

CREATE DATABASE lol_accs;
USE lol_accs;
GRANT ALL PRIVILEGES ON lol_accs.* TO 'DBUSER'@'localhost';

Then run db.sql to create the tables.
