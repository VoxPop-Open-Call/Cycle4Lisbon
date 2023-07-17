BINARY_NAME=main.out

build: docs compile

run: build
	./${BINARY_NAME}

test:
	go test -v -vet=all ./src/...

clean:
	go clean
	rm ${BINARY_NAME}

# Format code and `swag` annotations
format:
	swag fmt -d src/server/
	gofmt -w .

install:
	# Install swaggo/swag, necessary to build the docs
	go install github.com/swaggo/swag/cmd/swag@v1.16.1

.PHONY: docs
docs:
	swag init \
		--parseDependency \
		--markdownFiles . \
		-d src/server/,src/database/models,src/database/query,src/database/types,src/util/httputil \
		-g server.go

compile:
	go build -o ${BINARY_NAME} -v src/main.go
