.PHONY: vet test bench coverage

vet:
	go vet
	go vet -tags test_sha256

test:
	go test -v -cover -race
	go test -tags test_sha256 -v -cover -race

bench:
	go test -v -cover -test.bench=. -test.benchmem

coverage:
	go test -coverprofile=c.out && go tool cover -html=c.out && rm c.out
	go test -tags test_sha256 -coverprofile=c.out && go tool cover -html=c.out && rm c.out
