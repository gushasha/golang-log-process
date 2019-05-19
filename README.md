# golang-log-process

## Introduction
This is a mini demo for learning golang, it can process Nginx log to influxDb.

## Getting started
```
# clone the project
git clone https://github.com/gushasha/golang-log-process.git

# dependencies
You must install influxDb

# run
// In this case, you shoud set INFLUXDB_DSN in log_process.go
go run log_process.go
or
go run log_process.go -path ./access.log -influxDsn http://127.0.0.1:8086@test@testpassword@test@s

# mock data
go run mockdata.go
```