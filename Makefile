BUILD   := build/
TARGET  := $(BUILD)bd
DESTDIR := /usr/local/bin
GOFLAGS := -trimpath -buildmode=pie -mod=readonly -modcacherw -buildvcs=false

all: $(TARGET)

build: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.*
	go build $(GOFLAGS) -o $@ cmd/main.go

install:
	install -m755 $(TARGET) $(DESTDIR)/bd

clean:
	@rm -rf $(BUILD)

check:
	cat config.yaml | sed "s#~/#$(PWD)/$(BUILD)#g" > $(BUILD)config.yaml
	mkdir -p $(BUILD)bin $(BUILD)fs
	cd $(BUILD) && BD_CONFIG_FILE=$(PWD)/$(BUILD)config.yaml ./bd upgrade
