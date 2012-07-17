# Makefile to include for projects that are supposed to build debian
# packages
#
#   Packages the directory tree that is in the $DEB_ROOT directory (${PWD}/fakeroot
#   by default, can be overwritten) as a debian package. As an example, $DEB_ROOT/usr/bin/cmd
#   will be installed to /usr/bin/cmd when the debian package is installed.
#
#   It uses fpm (https://github.com/jordansissel/fpm/wiki) with sane defaults.
#   Pass extra options (dependencies, post-install scripts, etc.) by setting
#   the $DEB_ARGS variable.
#
#   Variables DEB_NAME, DEB_URL, DEB_VERSION, DEB_DESCRIPTION and DEB_MAINTAINER
#   must be set and the directory DEB_ROOT must exists (check it with 'make debcheck').
#
#   It automatically creates a .pkgignore file that includes everything apart from
#   the *.deb file.
#
#   It automatically does release book keeping using a file called .release, so
#   each build creates a different release (necessary for some apt repo software)
#
# usage:
#
#   build: debroot debbuild
#
#   DEB_NAME=myprog
#   DEB_URL=http://myprog.example.com
#   DEB_VERSION=$$(cat VERSION)
#   DEB_DESCRIPTION=MyProg is a super awesome program
#   DEB_MAINTAINER=Joe Plumber <joe.plumber@example.com>
#
#   # myprog depends on myprog-dep and myprog-dep2
#   DEB_ARGS=-d "myprog-dep" -d "myprog-deb2"
#
#   include deb.mk
#
#   # create the directory tree that gets packaged
#   debroot:
#     mkdir -p $(DEB_ROOT)/usr/bin
#     mkdir -p $(DEB_ROOT)/etc
#     cp bin/myprog $(DEB_ROOT)/usr/bin
#     cp myprog.conf $(DEB_ROOT)/etc/myprog.conf
#
#   clean: debclean
#
#
DEB_ROOT?=fakeroot
DEB_FPM_EXECUTABLE:=$$(dirname $$(dirname $$(gem which fpm)))/bin/fpm
RELEASE=$$(cat .release 2>/dev/null || echo "0")

debbuild: debcheck debclean bump_package_release
	$(DEB_FPM_EXECUTABLE) -n "$(DEB_NAME)" \
		-C "$(DEB_ROOT)" \
		--description "$(DEB_DESCRIPTION)" \
		$(DEB_ARGS) -v $(DEB_VERSION) -t deb -m "$(DEB_MAINTAINER)" --url "$(DEB_URL)" -s dir --iteration "$(RELEASE)" .;

	find . -mindepth 1 -maxdepth 1 | grep -v "\.deb" | sed 's/\.\///g' > .pkgignore


debcheck:
	@test -n "$(DEB_NAME)"        || { echo "**** Error: \$$DEB_NAME variable must be set to package name"                      && exit 1; }
	@test -n "$(DEB_URL)"         || { echo "**** Error: \$$DEB_URL variable must be set to the package's home page"            && exit 2; }
	@test -n "$(DEB_VERSION)"     || { echo "**** Error: \$$DEB_VERSION variable must be set to package's version"              && exit 3; }
	@test -n "$(DEB_DESCRIPTION)" || { echo "**** Error: \$$DEB_DESCRIPTION variable must be set to package's description"      && exit 4; }
	@test -n "$(DEB_MAINTAINER)"  || { echo "**** Error: \$$DEB_MAINTAINER variable must be set to the packagers email address" && exit 5; }
	@test -e "$(DEB_ROOT)"        || { echo "**** Error: \$$Fakeroot \"$(DEB_ROOT)\" does not exist -- nothing to package"      && exit 6; }
	@gem which fpm > /dev/null    || { echo "**** Error: can't find 'fpm' binary needed to build the package; install it with 'gem install fpm -v0.3.11'" && exit 7; }


debclean:
	rm -rf *.deb

bump_package_release:
		echo $$(( $(RELEASE) + 1 )) > .release
