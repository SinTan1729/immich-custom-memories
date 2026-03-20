PREFIX := /usr/local
PKGNAME := immich-custom-memories

build:
	go build -ldflags="-s -w" -o ${PKGNAME}

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

.PHONY: build install uninstall aur clean
