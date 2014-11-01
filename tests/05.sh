#! /bin/sh

. ./setup

setup_lishrc

NAME="allowed success via stdin"

begin

CMD="date +%Y"
CHECK=$(${CMD})
OUT=$(echo "${CMD}" | ${LISH})
if [ x"${OUT}" != x"${CHECK}" ]; then
	fail
fi

end

exit 0
