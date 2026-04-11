SHELL := $(firstword $(shell which bash sh))

modpkg := $(shell go list -f '{{.Name}}' .)
moddir := $(shell go list -f '{{.Dir}}' .)
modimp := $(shell go list -f '{{.ImportPath}}' .)

modsemver := $(shell cat "$(moddir)/VERSION" 2>/dev/null)
tagsemver := $(shell git describe --tags --abbrev=0 2>/dev/null)
newsemver := $(or $(tagsemver),$(modsemver))

output := dist
assets := README.md LICENSE

platforms := $(foreach os,linux darwin windows,$(foreach arch,amd64 arm64,$(os)-$(arch)))
platform  := $(or $(PLATFORM),$(platforms))

dist      := $(foreach p,$(platform),dist-$(p))
clean     := $(foreach p,$(platform),clean-$(p))
distclean := $(foreach p,$(platform),distclean-$(p))

os = $(firstword $(subst -, ,$(1)))
arch = $(lastword $(subst -, ,$(1)))

.PHONY: all generate version dist clean distclean bump-major bump-minor bump-patch $(platform) $(dist) $(clean) $(distclean)

all: $(platform)

# double-hyphen prevents usage from command-line.
# Make will interpret it as an invalid option and exit.
.PHONY: --force

# An empty recipe is always considered out of date.
# Any targets that depend on it will always be rebuilt.
--force:

# Bump the major, minor, or patch component of the version in VERSION.
# Uses `go run` to invoke the over command from the current project.
bump-major bump-minor bump-patch:
	@v=$$(go run . --$(subst bump-,,$@) < $(moddir)/VERSION) && echo "$$v" > $(moddir)/VERSION
	@echo "$(moddir) version $$(cat $(moddir)/VERSION)"

generate: version
	go generate -v ./...

version: $(moddir)/VERSION
	@echo "$(moddir) version $(shell cat $(moddir)/VERSION)"

dist: $(dist)

clean: $(clean)

distclean: $(distclean)

$(moddir)/VERSION: --force
ifeq ($(strip $(newsemver)),)
	$(error unknown version: set VERSION or tag the repository)
endif
	@echo ${newsemver} > $@

$(platform): generate
	@echo
	@echo build $@
	@echo
	@mkdir -p $(output)/$(modpkg)$(newsemver).$@
	GOOS=$(call os,$@) GOARCH=$(call arch,$@) go build -v -ldflags="-s -w" -o $(output)/$(modpkg)$(newsemver).$@/$(modpkg) .

.SECONDEXPANSION:
$(dist): $$(subst dist-,,$$@)
	@echo
	@echo dist $<
	@echo
	@cp $(assets) $(output)/$(modpkg)$(newsemver).$</
	tar -czf $(output)/$(modpkg)$(newsemver).$<.tar.gz -C $(output) $(modpkg)$(newsemver).$<

$(clean):
	@echo
	@echo clean $(subst clean-,,$@)
	@echo
	GOOS=$(call os,$(subst clean-,,$@)) GOARCH=$(call arch,$(subst clean-,,$@)) go clean -i -r $(modimp)

.SECONDEXPANSION:
$(distclean): $$(subst distclean-,clean-,$$@)
	@echo
	@echo distclean $(subst clean-,,$<)
	@echo
	rm -rf $(output)/$(modpkg)$(newsemver).$(subst clean-,,$<)
	rm -f $(output)/$(modpkg)$(newsemver).$(subst clean-,,$<).tar.gz
