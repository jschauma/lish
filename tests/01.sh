#! /bin/sh

. ./setup

setup_lishrc

NAME="allowed success via '-c'"

begin

CMD="date +%Y"
CHECK=$(${CMD})
OUT=$(${LISH} -c "${CMD}")
if [ x"${OUT}" != x"${CHECK}" ]; then
	fail
fi

end

exit 0
