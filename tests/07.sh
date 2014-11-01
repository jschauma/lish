#! /bin/sh

. ./setup

setup_lishrc

NAME="composite commands"

begin

CMD='du -h; date +%Y'
CHECK="$(eval ${CMD})"
OUT=$(${LISH} -c "${CMD}" 2>${STDERR})
if [ x"${OUT}" != x"${CHECK}" ]; then
	fail
fi

end

exit 0
