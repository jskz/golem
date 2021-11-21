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

https://user-images.githubusercontent.com/5122630/142783172-ff7281bc-9153-40c9-a839-81fd4970a2e6.mp4
