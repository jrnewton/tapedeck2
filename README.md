# tapedeck v2
This is a new version of [tapedeck](http://tapedeck.us) ([git repo](https://github.com/jrnewton/tapedeck)).  

It's currently a work in progress.  

You can see a [UI mockup](https://jrnewton.github.io/tapedeck2/) and try the [live server](https://new.tapedeck.us) (which is not always running).

## What
As a kid I spent a lot of time recording music with my tape deck from the huge variety of local college radio stations.  While college radio is alive and well thanks to streaming, archives come in differnet shapes and sizes and are limited to the last 2 weeks due to FCC rules.  This project provides a way to capture those archives for personal use and to record live streams when archives are not an option. 

## Why
I had fun making [tapedeck v1](https://github.com/jrnewton/tapedeck) but I never finished it.  It could only capture audio without a builtin playback mechanism.  Trying to revive the old codebase and infrastructure will take too much time due to a large number of dependencies and over engineering.  That was a known downside at the time as I wanted to experiment with a lot of AWS infra.  The fact that it's still running (albeit sans TLS) is pretty cool.  The goals for v2 are driven by these failings in v1.

## Goals
1. Simple architecture and design so it's easy to hack on.
2. Self-contained with limited dependencies. If the app works now, it should be able to continue working for years. Provide a package with a server binary and files for the web UI. If a database is needed, then use SQLite.
3. Need easy TLS support, as it's usually the hardest part when setting up new software. Consider punting with an nginx proxy and Let's Encrypt + certbot.

## TODO
- [x] TLS support.
- [ ] Auth support.
- [ ] Capture existing mp3 and m3u files (port existing functionality).
- [ ] Capture live streams.
- [ ] Plugin to capture existing archived shows from my favorite radio stations.
- [ ] Recordings must be shareable ala "anyone with this link can access".
- [ ] Playback GUI with offline support and background play on IOS (Eg doesn't shut off when you lock your phone).
- [ ] Data storage TBD.  It would work best for me to use Google drive but that goes against Goal #2.

## Dev Environment
- go version go1.23.4 linux/amd64
- Install package ```go get zombiezen.com/go/sqlite```
* If working under WSL, host your project under the native WSL filesystem and not under `/mnt`:
```
Under WSL2, anything under /mnt/*/ is like a remote filesystem, that goes over the Plan 9 Filesystem Protocol (9P).

There sqlite can't lock it's database files. Everywhere else, outside the shared filesystem, sqlite will work.
```
Via https://github.com/microsoft/WSL/issues/2395#issuecomment-909045977
