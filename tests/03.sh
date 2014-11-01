#! /bin/sh

. ./setup

setup_lishrc

NAME="allowed success via SSH_ORIGINAL_COMMAND overriding '-c'"

begin

CMD="date +%Y"
CHECK=$(${CMD})
OUT=$(SSH_ORIGINAL_COMMAND="${CMD}" ${LISH} -c "ls")
if [ x"${OUT}" != x"${CHECK}" ]; then
	fail
fi

end

exit 0
