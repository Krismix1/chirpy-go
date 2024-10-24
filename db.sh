#!/usr/bin/env sh

goose postgres "user=postgres dbname=chirpy sslmode=disable" -dir sql/schema $1
