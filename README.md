# tapedeck v2
This is a new version of [tapedeck](http://old.tapedeck.us) ([git repo](https://github.com/jrnewton/tapedeck)).  

:construction: This is a work in progress :construction:

You can see a [UI mockup](https://jrnewton.github.io/tapedeck2-mockups/) and try the [live server](https://tapedeck.us).

## What
As a kid I spent a lot of time recording music with my tape deck from the huge variety of local community radio stations.  While this type of radio is alive and well thanks to streaming, their archives come in differnet shapes and sizes and are limited in time due to FCC rules.  This project provides a way to capture those archives for personal use and to record live streams when archives are not an option.

## Why v2
I had fun making [tapedeck v1](https://github.com/jrnewton/tapedeck) but I never finished it.  It could only capture audio without a builtin playback mechanism.  Trying to revive the old codebase and infrastructure will take too much time due to a large number of dependencies and over engineering.  That was a known downside at the time as I wanted to experiment with a lot of AWS infra.  The fact that it's still running (albeit sans TLS) is pretty cool.  The goals for v2 are driven by these failings in v1.

## Goals
1. Simple architecture and design so it's easy to hack on.
2. Self-contained with limited dependencies. If the app works now, it should be able to continue working for years. Provide a package with a server binary and files for the web UI. If a database is needed, then use SQLite.
3. Need easy TLS support, as it's usually the hardest part when setting up new software.

## TODO
- [x] TLS support, via [nginx](config/prod/default.nginx).
- [x] Auth support, via nginx and [oauth2-proxy](config/prod/oauth2-proxy.toml).
- [x] Production deployment, digital ocean droplet with [rsync deploy](Makefile#L56) and [systemd units](config/prod).
- [ ] Capture existing mp3 and m3u files (port existing functionality).
- [ ] Capture existing archived shows from my favorite radio stations.
- [ ] Recordings must be shareable ala "anyone with this link can access".
- [ ] Playback GUI with offline support and background play on IOS (Eg doesn't shut off when you lock your phone).
- [ ] Data storage TBD.  It would work best for me to use Google drive but that goes against Goal #2.
- [ ] Capture live streams.

## Local Dev
- go version go1.23.4 linux/amd64
- Install package `go get zombiezen.com/go/sqlite`
- Install package `go get github.com/google/uuid`
- If working under WSL, host your project under the native WSL filesystem and not under `/mnt` [source](https://github.com/microsoft/WSL/issues/2395#issuecomment-909045977)
```
Under WSL2, anything under /mnt/*/ is like a remote filesystem, that goes over the Plan 9 Filesystem Protocol (9P).

There sqlite can't lock it's database files. Everywhere else, outside the shared filesystem, sqlite will work.
```
- Use [launch.json](config/dev/launch.json) to launch the server locally.

## Prod Environment
### Install tapedeck
- install [tapedeck.service](config/prod/tapedeck.service) following instructions in the file.
- `make upload`

### nginx and certbot
- install nginx and certbot
```
sudo apt install nginx certbot python3-certbot-nginx
systemctl enable nginx
```
- install the certs
```
certbot --nginx -d tapedeck.us --agree-tos
```
- test config and restart `nginx -t && systemctl restart nginx`
- visit your site to verify nginx works, certs valid

### oauth2-proxy
- install oauth2-proxy binary to `/usr/local/sbin/`
- generate a cookie secret. See https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview
- configure Google as oauth provider: https://oauth2-proxy.github.io/oauth2-proxy/configuration/providers/google
- copy [oauth2-proxy.toml](config/prod/oauth2-proxy.toml) to `/etc/oauth2-proxy/oauth2-proxy.toml`
- install [oauth2-proxy.service](config/prod/oauth2-proxy.service) following instructions in the file.

### Update nginx with oauth2-proxy
- update `/etc/nginx/sites-available/default` with [default](config/prod/default.nginx) to enable oauth2-proxy

### Restart nginx
- test config changes and restart `nginx -t && systemctl restart nginx`
