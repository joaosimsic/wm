.PHONY: build run clean fmt vet

BIN      = ./wm
XEPHYR   = Xephyr
XDISPLAY = :1
XRES     = 1920x1080

build:
	go build -o $(BIN) ./cmd/wm/

run: build
	@which $(XEPHYR) >/dev/null 2>&1 || { echo "Xephyr not found. Install: sudo apt install xserver-xephyr"; exit 1; }
	@rm -f /tmp/.X$(subst :,,$(XDISPLAY))-lock
	@echo "Xephyr on $(XDISPLAY) ($(XRES))..."
	$(XEPHYR) -br -ac -noreset -screen $(XRES) $(XDISPLAY) &
	@sleep 1
	DISPLAY=$(XDISPLAY) $(BIN); kill %1 2>/dev/null; rm -f /tmp/.X$(subst :,,$(XDISPLAY))-lock

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f $(BIN)
