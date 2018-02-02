GOPATH := ${PWD}
export GOPATH

clean:
	rm access.log examples/main

build:
	cd examples && go build -o main

test:
	make testw3chttpd
	make testmonitor

testw3chttpd:
	cd src/w3chttpd && go test -v -race

testmonitor:
	cd src/monitor && go test -v -race

benchmark:
	make benchw3chttpd
	make benchmonitor

benchw3chttpd:
	cd src/w3chttpd && go test -run=NONE -bench=. -benchmem -race

benchmonitor:
	cd src/monitor && go test -run=NONE -bench=. -benchmem -race

run:
	./examples/main


