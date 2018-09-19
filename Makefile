VERSION := $(shell cat ./VERSION)

linux:
	GOOS=linux GOARCH=amd64 go build -o releases/linux_go_shoeshop_hw .

windows:
	GOOS=windows GOARCH=amd64 go build -o releases/win_go_shoeshop_hw.exe .

raspberry:
	GOOS=linux GOARCH=arm GOARM=6 go build -o releases/arm_go_shoeshop_hw .

osx:
	GOOS=darwin GOARCH=amd64 go build -o releases/osx_go_shoeshop_hw .

install:
	go install .

release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)

.PHONY: windows linux raspberry osx release