# Golem

## Overview

Golem is a from-scratch attempt at a Diku-like MUD implemented with [Go](https://golang.org/) in 2021.

## Docker-based Setup

## Retrieve, Build, Start

```
git clone git@github.com:jskz/golem.git
cd golem
docker build --tag golem:latest .
docker-compose up
```

The MUD is exposed on the host's TCP port 4000 by default.

A phpMyAdmin instance is exposed on port 8000 providing root access to the game's MySQL storage.

## Destroying all database data and starting over

```
docker-compose down
docker volume rm golem_db_data
```

## Video

https://user-images.githubusercontent.com/5122630/144722737-9cff11c9-9127-4d8a-a075-b32e82b7aaf5.mp4

## Notes

- Until 1.0, I'm playing fast and loose with the `migrations` directory and regularly editing old files.  After 1.0, the expected incremental migration with descriptive snake-case filename structure will be honoured.
