PREFIX := /usr/local
PKGNAME := immich-custom-memories
GIT_VERSION := $(shell git describe --tags --abbrev=0)

build:
	go build -ldflags="-s -w -X 'main.Version=${GIT_VERSION}'" -o ${PKGNAME}

install: build
	install -Dm755 $(PKGNAME) "$(DESTDIR)$(PREFIX)/bin/$(PKGNAME)"
	install -Dm644 $(PKGNAME).1 "$(DESTDIR)$(PREFIX)/man/man1/$(PKGNAME).1"

uninstall:
	rm -f "$(DESTDIR)$(PREFIX)/bin/$(PKGNAME)"
	rm -f "$(DESTDIR)$(PREFIX)/man/man1/$(PKGNAME).1"

aur: build 
	tar --transform 's/.*\///g' -czf $(PKGNAME).tar.gz $(PKGNAME) $(PKGNAME).1
 
clean:
	rm -f "${PKGNAME}"
	rm -f "${PKGNAME}.tar.gz"

GH_TOKEN := $(shell cat ~/.config/github_token)
release: aur
	gh release create "${GIT_VERSION}" --notes "$$(git-cliff --latest --github-token ${GH_TOKEN})" "$(PKGNAME).tar.gz"
	$(MAKE) clean

.PHONY: build install uninstall aur clean release
