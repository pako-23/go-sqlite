TESTS := sqlite_test.go
SRCS := sqlite.go \
	sqlite_vtable.go \
	sqlite3.c \
	sqlite3.h \
	sqlite3ext.h

.PHONY: all build check clean coverage

all: build

build: $(SRCS)
	go build

check: $(SRCS) $(TESTS)
	go test -covermode=atomic -race

coverage: coverage.html

coverage.cov: $(SRCS) $(TESTS)
	go test -covermode=atomic -race -coverprofile=$@

coverage.html: coverage.cov
	go tool cover -html=$^ -o $@

clean:
	- rm -f coverage.html
	- rm -f coverage.cov
