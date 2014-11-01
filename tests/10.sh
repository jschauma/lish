#! /bin/sh

. ./setup

setup_lishrc

NAME="cd builtin"

begin

CMD='cd /; pwd; cd; pwd;'
CHECK="$(eval ${CMD})"
OUT=$(${LISH} -c "${CMD}" 2>${STDERR})
if [ x"${OUT}" != x"${CHECK}" ]; then
	fail
fi

end

exit 0
