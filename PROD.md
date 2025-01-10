# Config of tapedeck.us

## Install tapedeck
- install [tapedeck.service](config/tapedeck.service) following instructions in the file.
- `make upload`

## nginx and certbot
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

## oauth2-proxy
- install oauth2-proxy binary to `/usr/local/sbin/`
- generate a cookie secret. See https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview
- configure Google as oauth provider: https://oauth2-proxy.github.io/oauth2-proxy/configuration/providers/google
- copy [oauth2-proxy.toml](config/oauth2-proxy.toml) to `/etc/oauth2-proxy/oauth2-proxy.toml`
- install [oauth2-proxy.service](config/oauth2-proxy.service) following instructions in the file.

## Update nginx with oauth2-proxy
- update `/etc/nginx/sites-available/default` with [default](config/default.nginx) to enable oauth2-proxy

## Restart nginx
- test config changes and restart `nginx -t && systemctl restart nginx`
