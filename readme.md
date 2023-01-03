# countInLogs

The application counts number of regex strings (`search` parameter) in all log files fount in: `logs_dir` parameter, and then filter results using data from file: `count_from_file`.
Serach of ID numbers in ~50Mb logs takes ~1.5sec
Serach of ID numbers in ~30GB logs takes ~10min

NOTE: some ideas was used from the article: https://medium.com/swlh/processing-16gb-file-in-seconds-go-lang-3982c235dfa2 published by Ohm Patel - the article provides better approach for searching in logs by date/time


## Build

if your you're building in the same platform you can use simple:
```sh
go build -o ./build .
```
for cross platform building you can use:
```sh
env GOOS=target-OS GOARCH=target-architecture go build 
```
more details in: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04


### Cross platform: Build for Linux
```sh
env GOOS=linux GOARCH=amd64 go build -o ./build .
```
### Cross platform: Build for Windows
```sh
env GOOS=windows GOARCH=386 go build -o ./build/countInLogs.exe .
```

## Execution
for Windows:
```sh
countInLogs.exe --count_from_file .\Ids.txt --logs_dir .\log_files -greater_than 100
```
