#! /bin/sh

. ./setup

setup_lishrc

NAME="allowed failure"

begin

CMD="du -?"
${LISH} -c "${CMD}" >/dev/null 2>${STDERR}
if [ $? -ne 64 ]; then
	fail
fi

exit 0
