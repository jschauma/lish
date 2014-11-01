#! /bin/sh

. ./setup

setup_lishrc

NAME="forbidden commands"

begin

CMD='ls'
${LISH} -c "${CMD}" 2>${STDERR}
if [ $? -ne 127 ]; then
	fail
fi

end

exit 0
