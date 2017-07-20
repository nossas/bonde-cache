# Bonde Cache

## Features

1. http web server with maximum performance
2. read json from mobilizations
3. parse html to save in bolt at startup script
4. configure bolt as in-memory db
5. serve domains based on custom_domain in mobilizations
6. add worker to listen queue and update bolt cache from mobilization
7. add support to auto tls custom_domain
8. production test with multiple domains and certificates
9. check if request is not empty then not override cache saved previous content
10. enabled grateful shutdown when server is restarted because new domains was added
11. document how to configure persistent storage using dokku to save certificates and db to local file
12. create script to sync local files (db and certificates) to s3
12. create terraform to load ec2 instance
13. create shell script to setup instance and configure docker to run bonde-cache
14. save to analytics db the views and stats from each request recieved
15. provide some stats by host

As ready as possible to shutdown the lights and close the door.

## Development
### Docker support

```
docker build -t nossas/bonde-cache .
docker run -it --rm -p 443:443 -v "$PWD":/go/src/app -w /go/src/app -e CACHE_PORT=80 -e PORT=80 -e CACHE_PORTSSL=443 -e CACHE_INTERVAL=20 -e CACHE_RESET=false -e CACHE_ENV=development -e AWS_ACCESS_KEY_ID= -e AWS_SECRET_ACCESS_KEY=  --name bonde-cache-app nossas/bonde-cache app
```

### Generate self-signed ssl

```
openssl req -new -newkey rsa:2048 -sha1 -days 3650 -nodes -x509 -subj "/C=US/ST=Georgia/L=Atlanta/O=BNR/CN=www.en.nossas.org" -keyout server.key -out server.crt
```

### Starting

Add do hosts ```127.0.0.1    www.en.nossas.org``` and try to access the url at browser.


## Production

### Dokku support

```
dokku apps:create 00-cache

dokku config:set 00-cache CACHE_ENV=production
dokku config:set 00-cache CACHE_INTERVAL=40
dokku config:set 00-cache CACHE_PORT=80
dokku config:set 00-cache CACHE_PORTSSL=443
dokku config:set 00-cache CACHE_RESET=false
dokku config:set 00-cache DOKKU_DOCKERFILE_PORTS="443/tcp 80/tcp"
dokku config:set 00-cache DOKKU_PROXY_PORT_MAP="http:443:443 http:80:80"
dokku config:set 00-cache PORT=80
dokku config:set 00-cache AWS_SECRET_ACCESS_KEY=
dokku config:set 00-cache AWS_ACCESS_KEY_ID=

dokku storage:mount 00-cache /var/lib/dokku/data/storage/cache-certificates:/go/src/app/data/certificates
dokku storage:mount 00-cache /var/lib/dokku/data/storage/cache-db:/go/src/app/data/db

```