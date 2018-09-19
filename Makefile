VERSION := $(shell cat ./VERSION)

linux:
	GOOS=linux GOARCH=amd64 go build ...

windows:
	GOOS=windows GOARCH=amd64 go build ...

raspberry:
	GOOS=linux GOARCH=arm GOARM=6 go build ...

osx:
	GOOS=darwin GOARCH=amd64 go build ...

install:
	go install .

release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)

.PHONY: windows linux raspberry osx release