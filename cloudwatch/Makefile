all : dep build

dep:
	dep ensure

build: dep
	GOOS=linux GOARCH=amd64 go build -o cloudwatch main.go
	zip cloudwatch.zip cloudwatch
	mkdir -p build
	cp cloudwatch.zip build
	cp sam-cloudwatch-logs.yml build