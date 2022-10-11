#!/bin/bash

go build -o go-course cmd/web/*.go && ./go-course -dbname=go-course -dbuser=postgres -dbpass=postgres -cache=false -prod=false