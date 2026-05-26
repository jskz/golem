# Golem

## Overview

Golem is a from-scratch attempt at a Diku-like MUD implemented with [Go](https://golang.org/) in 2021.

## Build Status

[![master CI status](https://github.com/jskz/golem/actions/workflows/build-and-push-image.yml/badge.svg?branch=master)](https://github.com/jskz/golem/actions/workflows/build-and-push-image.yml)

## Docker-based Setup

## Retrieve, Build, Start

```
git clone git@github.com:jskz/golem.git
cd golem
docker build --tag golem:latest .
docker-compose up
```

The MUD is exposed on the host's TCP port 4000 by default.

## Destroying all database data and starting over

```
docker-compose down
docker volume rm golem_db_data
```

## Video

https://user-images.githubusercontent.com/5122630/144722737-9cff11c9-9127-4d8a-a075-b32e82b7aaf5.mp4
