NAME=lish
VERSION=$(shell sed -n -e 's/^const VERSION = "\(.*\)"/\1/p' src/${NAME}.go)
RPMREV=$(shell awk '/%define release/ { print $$3;}' rpm/${NAME}.spec.in)
UNAME=$(shell uname)

.PHONY: test

PREFIX?=/usr/local

help:
	@echo "The following targets are available:"
	@echo "build      build the executable"
	@echo "clean      remove temporary build files"
	@echo "install    install ${NAME} into ${PREFIX}"
	@echo "rpm        build an RPM of ${NAME}-${VERSION}-${RPMREV} on ${BUILDHOST}"
	@echo "sign       sign the RPM package"
	@echo "test       run all tests under tests/"
	@echo "uninstall  uninstall ${NAME} from ${PREFIX}"

rpm: spec buildrpm

spec: rpm/${NAME}.spec

rpm/${NAME}.spec: rpm/${NAME}.spec.in
	cat $< CHANGES | sed -e "s/VERSION/${VERSION}/" >$@

build: src/${NAME}

src/${NAME}: src/${NAME}.go
	if [ ${UNAME} = "Darwin" ]; then		\
		CC=clang;				\
	fi;						\
	CC=$${CC} go build -o src/${NAME} src/${NAME}.go

buildrpm: packages/rpms/${NAME}-${VERSION}-${RPMREV}.x86_64.rpm

linux_binary:
	ssh ${BUILDHOST} "mkdir -p ${NAME}"
	rsync -e ssh -avz --exclude packages/ --exclude .git/ . ${BUILDHOST}:${NAME}/.
	ssh ${BUILDHOST} "cd ${NAME}/src && rm -f ${NAME} && GOROOT=~/go ~/go/bin/go build ${NAME}.go"

packages/rpms/${NAME}-${VERSION}-${RPMREV}.x86_64.rpm: spec linux_binary
	ssh ${BUILDHOST} "cd ${NAME}/rpm && sh mkrpm.sh ${NAME}.spec"
	scp ${BUILDHOST}:redhat/RPMS/*/${NAME}-${VERSION}-${RPMREV}.x86_64.rpm packages/rpms/
	ls packages/rpms/${NAME}-${VERSION}-${RPMREV}.x86_64.rpm;

sign: packages/rpms/${NAME}-${VERSION}-${RPMREV}.x86_64.rpm.asc

packages/rpms/${NAME}-${VERSION}-${RPMREV}.x86_64.rpm.asc: packages/rpms/${NAME}-${VERSION}-${RPMREV}.x86_64.rpm
	gpg -b -a packages/rpms/${NAME}-${VERSION}-${RPMREV}.x86_64.rpm

version:
	echo ${VERSION} > packages/version

install: build
	mkdir -p ${PREFIX}/bin ${PREFIX}/share/man/man1
	install -c -m 0555 src/${NAME} ${PREFIX}/bin/${NAME}
	install -c -m 0555 doc/${NAME}.1 ${PREFIX}/share/man/man1/${NAME}.1

uninstall:
	rm -f ${PREFIX}/bin/${NAME} ${PREFIX}/share/man/man1/${NAME}.1

test: src/${NAME}
	@cd tests && for t in *.sh; do			\
		sh $${t};				\
	done

clean:
	rm -f src/${NAME} rpm/${NAME}.spec
