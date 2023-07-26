#!/bin/bash
go build -o oseeea cmd/web/*.go
./oseeea -dbname=oseeea.go -dbuser=postgres -dbpass=Science@1992 -cache=false -production=false