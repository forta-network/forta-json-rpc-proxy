.PHONY: fmt-sol
fmt-sol:
	forge fmt testing/contracts/

.PHONY: testproxy
testproxy:
	go run testing/testproxy/main.go
