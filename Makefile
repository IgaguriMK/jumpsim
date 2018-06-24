.PHONY: build
build: jumpsim sphereDensity

.PHONY: jumpsim
jumpsim:
	go build jumpsim.go


.PHONY: sphereDensity
sphereDensity:
	go build sphereDensity.go
