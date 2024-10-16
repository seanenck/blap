BUILD   := build/
OBJECT  := blap
TARGET  := $(BUILD)$(OBJECT)
DESTDIR := /usr/local/bin
GOFLAGS := -trimpath -buildmode=pie -mod=readonly -modcacherw -buildvcs=false
VERSION ?= $(shell git log -n 1 --format=%h)

all: $(TARGET)

build: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.*
	go build $(GOFLAGS) -ldflags "$(LDFLAGS) -X main.version=$(VERSION)" -o $@ cmd/main.go

install:
	install -m755 $(TARGET) $(DESTDIR)/$(OBJECT)

clean:
	@rm -rf $(BUILD)

unittest:
	go test -v ./...

check: unittest $(TARGET)
	cat config.yaml | sed "s#~/#$(PWD)/$(BUILD)#g" > $(BUILD)config.yaml
	mkdir -p $(BUILD)bin $(BUILD)fs
	cd $(BUILD) && BLAP_CONFIG_FILE=$(PWD)/$(BUILD)config.yaml ./$(OBJECT) upgrade
