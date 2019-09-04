# LAM - LoL Account Manager

A Webapp to store and share League of Legends accounts.

# Install

After installing Go and configuring $GOPATH, run:

```
go get github.com/erikfastermann/lam
```

# Docker

After installing Docker and Docker-compose, run:

```
sudo docker-compose up -d
```

# Add a User

Add a user with the manage.go script by running:

```
cd $GOPATH/src/github.com/erikfastermann/lam/manage
go get
go run manage.go -a $LAM_DB_PATH
```

If you don't supply a path, the default for this docker volume will be used.

You might need sudo for this to work.

# HTTPS

To activate HTTPS, comment out and adjust the following in the docker-compose.yml file:

```yaml
- path/to/keypairs:/var/lam/keypairs
```

```yaml
LAM_HTTPS_DOMAIN: 'https://your-domain.com'
LAM_HTTPS_ADDRESS: ':443'
LAM_HTTPS_CERTS: '/var/lam/keypairs/cert-file1,/var/lam/keypairs/cert-file2'
LAM_HTTPS_KEYS: '/var/lam/keypairs/key-file1,/var/lam/keypairs/key-file2'
```

NOTE: Don't use different ports for inside and outside the container.
Redirecting to HTTPS will not work otherwise.

# Appendix

## List of environment variables

Address (e.g.: ':80'): `LAM_ADDRESS`

Template Glob (e.g.: 'template/*'): `LAM_TEMPLATE_GLOB`

Sqlite3 DB Path (e.g.: '/lam.db'): `LAM_DB_PATH`

### HTTPS

Domain (used to upgrade from HTTP to HTTPS): `LAM_HTTPS_DOMAIN`

Address (set to activate): `LAM_HTTPS_ADDRESS`

Cert files (eg.: 'path/to/cert-file,path/to/another/cert-file'): `LAM_HTTPS_CERTS`

Key files (eg.: 'path/to/key-file,path/to/another/key-file'): `LAM_HTTPS_KEYS`
