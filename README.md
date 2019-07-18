# Dadata exporter
[![Build Status](https://travis-ci.org/StupidScience/dadata-exporter.svg?branch=master)](https://travis-ci.org/StupidScience/dadata-exporter)
[![Coverage Status](https://coveralls.io/repos/github/StupidScience/dadata-exporter/badge.svg)](https://coveralls.io/github/StupidScience/dadata-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/StupidScience/dadata-exporter)](https://goreportcard.com/report/github.com/StupidScience/dadata-exporter)

Get account statistics for [dadata service API](https://dadata.ru/api/) and expose it in prometheus format.

## Configuration

Exporter configurates via environment variables:

|Env var|Description|
|---|---|
|DADATA_TOKEN|Authorization token for access to Dadata API|
|DADATA_X_SECRET|X-Secret header for access to Dadata Api|

Exporter listen on tcp-port **9501**. Metrics available on `/metrics` path.

## Exposed metrics

|Metric|Descrpition|
|---|---|
|dadata_current_balance|Current balance on Dadata|
|dadata_services_clean_total|Clean count for today|
|dadata_services_merging_total|Merging count for today|
|dadata_services_suggestions_total|Suggestions count for today|

## Run via Docker

The latest release is automatically published to the [Docker registry](https://hub.docker.com/r/stupidscience/nodeping-exporter).

You can run it like this:
```
$ docker run -d --name dadata-exporter \
            -e DADATA_TOKEN=12345abcdef \
            -e DADATA_X_SECRET=12345abcdef \
            -p 9501:9501 \
            stupidscience/dadata-exporter
```
