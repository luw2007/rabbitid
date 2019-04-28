# rabbitid
TAG=$(shell git tag | sort -bt"." -k1,1 -k2,2n -k3,3n |tail -n 1)
GO_IMAGES=rabbitid:${TAG}

fmt:
	find . -name "*.go" -not -path "./vendor/*" -not -path ".git/*" | xargs gofmt -w -s -d
	find . -name "*.go" -not -path "./vendor/*" -not -path ".git/*" | xargs goimports -w
	find . -type f -not -path "./vendor/*" -not -path "./.idea/*" -not -path "./.git/*" -print0 | xargs -0 misspell
http:
	go run cmd/idHttp/main.go
redis:
	go run cmd/idRedis/main.go
wrk:
	wrk -c 10 -t 2 -d 5 http://127.0.0.1:7000/next -s tools/wrk.lua
test:
	docker-compose up -d
	go test github.com/luw2007/rabbitid/generator -bench . -benchmem
	go test github.com/luw2007/rabbitid/store
	go test github.com/luw2007/rabbitid/cmd/idHttp/service -bench . -benchmem
	docker-compose down
build:
	docker build -t ${GO_IMAGES} .
