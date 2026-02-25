PACKAGE=vdu
PACKAGE_PREFIX=$(PACKAGE)-
VERSION := $(shell grep '__version__ = ' vdu.py | cut -d'"' -f2)
RELEASE=$(PACKAGE)-$(VERSION)
RELEASE_DIR=$(RELEASE)
DEBIAN_FILE=./$(RELEASE).deb

default: 
	@echo "=== Instructions for Debian package $(PACKAGE) v$(VERSION) ==="
	@echo ""
	@echo "Available commands:"
	@echo "  make env       - Create .venv and install dependencies (requirements.txt)"
	@echo "  make test      - Run pytest tests"
	@echo "  make build     - Build .deb with dpkg-deb --build $(RELEASE_DIR)"
	@echo "  make install   - Install: sudo apt install $(DEBIAN_FILE)"
	@echo "  make uninstall - Uninstall: sudo apt remove $(PACKAGE)"
	@echo ""
	@echo "Typical workflow:"
	@echo "  1. make env"
	@echo "  2. make test"
	@echo "  3. Prepare $(RELEASE_DIR)/debian/control, rules..."
	@echo "  4. make build"
	@echo "  5. make install"
	@echo ""
	@echo "Parameters:"
	@echo "  LANG=<python|go(default)>"
	@echo ""

env:
	python -m venv .venv
	. .venv/bin/activate &&	pip install -r requirements.txt

test: 
ifeq ($(LANG),python)
	$(MAKE) test-python
else
	$(MAKE) test-go
endif

compile:
	@cd go && go build

test-go: compile
	@echo "Running tests..."
	@mkdir -p ~/tmp
	cd go && TMPDIR=~/tmp go test
	
test-python: env
	@echo "Running tests..."
	. .venv/bin/activate && pytest

prepare: 
ifeq ($(LANG),python)
	$(MAKE) prepare-python
else
	$(MAKE) prepare-go
endif

prepare-python: test-python
	rm -rf $(PACKAGE_PREFIX)*
	mkdir -p $(RELEASE_DIR)/DEBIAN $(RELEASE_DIR)/usr/bin $(RELEASE_DIR)/usr/lib/python3/dist-packages/vdu
	sed "s/VERSION/$(VERSION)/g" control.python > $(RELEASE_DIR)/DEBIAN/control
	cp vdu $(RELEASE_DIR)/usr/bin/
	cp vdu.py __init__.py $(RELEASE_DIR)/usr/lib/python3/dist-packages/vdu

prepare-go: test-go
	rm -rf $(PACKAGE_PREFIX)*
	mkdir -p $(RELEASE_DIR)/DEBIAN $(RELEASE_DIR)/usr/bin
	sed "s/VERSION/$(VERSION)/g" control.go > $(RELEASE_DIR)/DEBIAN/control
	cp go/vdu $(RELEASE_DIR)/usr/bin/

build: prepare
	dpkg-deb --build $(RELEASE_DIR)

install: build
	sudo apt install -y $(DEBIAN_FILE)

uninstall:
	sudo apt remove -y $(PACKAGE)
