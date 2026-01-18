PACKAGE=vdu
VERSION=0.1.0
RELEASE=$(PACKAGE)-$(VERSION)
RELEASE_DIR=$(RELEASE)
DEBIAN_FILE=./$(RELEASE).deb

default: 
	@echo "=== Instructions for Debian package $(PACKAGE) v$(VERSION) ==="
	@echo ""
	@echo "Available commands:"
	@echo "  make env     - Create .venv and install dependencies (requirements.txt)"
	@echo "  make test    - Run pytest tests"
	@echo "  make build   - Build .deb with dpkg-deb --build $(RELEASE_DIR)"
	@echo "  make install - Install: sudo apt install $(DEBIAN_FILE)"
	@echo "  make remove  - Remove: sudo apt remove $(PACKAGE)"
	@echo ""
	@echo "Typical workflow:"
	@echo "  1. make env"
	@echo "  2. make test"
	@echo "  3. Prepare $(RELEASE_DIR)/debian/control, rules..."
	@echo "  4. make build"
	@echo "  5. make install"
	@echo ""
# 	@echo "Note: Make sure $(RELEASE_DIR)/debian/ is properly configured!"

env:
	python -m venv .venv
	. .venv/bin/activate &&	pip install -r requirements.txt

test:
	pytest

build:
	dpkg-deb --build $(RELEASE_DIR)

install:
	sudo apt install -y $(DEBIAN_FILE)

remove:
	sudo apt remove -y $(PACKAGE)
