.PHONY: all build clean

OUT_DIR := out

PKGS := $(shell find . -name "main.go" -not -path "*/vendor/*" -not -path "*/.*" | xargs -n1 dirname | sort -u | sed 's|^\./||')

BINS := $(addprefix $(OUT_DIR)/,$(PKGS))

all: build

build: $(BINS)

$(BINS): $(OUT_DIR)/%: %/main.go
	@mkdir -p $(OUT_DIR)
	$(eval PACKAGE := $(notdir $*))
	cd $* && go build -o ../$(OUT_DIR)/$(PACKAGE) .

clean:
	rm -rf $(OUT_DIR)
