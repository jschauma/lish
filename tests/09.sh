#! /bin/sh

. ./setup

setup_lishrc

NAME="combined commands, forbidden and allowed"

begin

CMD='ls; date +%Y'
CHECK="$(date +%Y)"
OUT=$(${LISH} -c "${CMD}" 2>${STDERR})
if [ x"${OUT}" != x"${CHECK}" ]; then
	fail
fi

end

exit 0
