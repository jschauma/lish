#! /bin/sh

. ./setup

setup_lishrc

NAME="'-c' overrides stdin"

begin

CMD="date +%Y"
CHECK=$(${CMD})
OUT=$(echo "ls" | ${LISH} -c "${CMD}")
if [ x"${OUT}" != x"${CHECK}" ]; then
	fail
fi

end

exit 0
