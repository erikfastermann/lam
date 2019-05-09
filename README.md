# League Accs

## Environment vars

User names: LEAGUE_ACCS_USER1 LEAGUE_ACCS_USER2
User passwords: LEAGUE_ACCS_USER1_PW LEAGUE_ACCS_USER2_PW
CSRF Token (32 Byte): LEAGUE_ACCS_CSRF
Template Dir (e.g.: /tmp/*): LEAGUE_ACCS_TEMPLATE_DIR
Accounts Json File: LEAGUE_ACCS_JSON

Setting all of them is mandatory for a usable experience.

## JSON Format

```
[
  {
    "region": "euw",
    "tags": [ 
        "Awesome",
        "Cool"
    ],
    "ign": "in-game-name",
    "username": "username",
    "password": "12345678",
    "user": "LEAGUE_ACCS_USER1",
    "leaverbuster": "",
    "ban": "permanent",
    "ban_recently": "",
    "owner_active": true,
    "password_changed": true,
    "pre_30": true,
    "elo": ""
  },
  {
    "region": "ru",
    "tags": [],
    "ign": "something",
    "username": "something",
    "password": "abcdef123",
    "user": "LEAGUE_ACCS_USER2",
    "leaverbuster": "5 min",
    "ban": "1970-01-01 11:11",
    "ban_recently": "1970-01-01 10:10",
    "owner_active": false,
    "password_changed": false,
    "pre_30": false,
    "elo": "Silber 1"
  }
]
```
