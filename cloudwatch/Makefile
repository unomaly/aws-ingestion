all : dep build

dep:
	dep ensure

build: dep
	GOOS=linux GOARCH=amd64 go build -o cloudwatch main.go
	zip cloudwatch.zip cloudwatch
	mkdir -p cloudwatch-0.1.0
	cp cloudwatch.zip cloudwatch-0.1.0
	cp sam-cloudwatch-logs.yml cloudwatch-0.1.0/sam-cloudwatch-logs-0.1.0.yml