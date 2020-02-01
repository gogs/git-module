.PHONY: vet test bench coverage

vet:
	go vet

test:
	go test -v -cover -race

bench:
	go test -v -cover -test.bench=. -test.benchmem

coverage:
	go test -coverprofile=c.out && go tool cover -html=c.out && rm c.out
