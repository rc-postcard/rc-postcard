GO           ?= go
BIN           = rc-postcard
SRC           = $(shell find . -type f -name '*.go')
.DEFAULT_GOAL = build

build: $(BIN)

run: $(BIN)
	./$<

$(BIN): $(SRC) go.mod go.sum home.html address.gohtml no-address-home.html
	$(GO) build

clean:
	$(RM) $(BIN)

.PHONY: test
test:
	$(GO) test ./...

.PHONY: pg
pg:
	docker run --rm --name rc-postcard-pg -d \
		-e POSTGRES_DB=metadata -e POSTGRES_HOST_AUTH_METHOD=trust \
		-p 5432:5432 postgres