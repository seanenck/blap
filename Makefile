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
