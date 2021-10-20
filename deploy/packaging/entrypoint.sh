#!/usr/bin/env sh


set -e

# check to see if this file is being run or sourced from another script
_is_sourced() {
	# https://unix.stackexchange.com/a/215279
	[ "${#FUNCNAME[@]}" -ge 2 ] \
		&& [ "${FUNCNAME[0]}" = '_is_sourced' ] \
		&& [ "${FUNCNAME[1]}" = 'source' ]
}

# check arguments for an option that would cause /__PROJECT__ to stop
# return true if there is one
_want_help() {
	local arg
	for arg; do
		case "$arg" in
			-'?'|--help|-v)
				return 0
				;;
		esac
	done
	return 1
}

_main() {
	# if command starts with an option, prepend __PROJECT__
	if [ "${1:0:1}" = '-' ]; then
		set -- /__PROJECT__ "$@"
	fi
		# skip setup if they aren't running /__PROJECT__ or want an option that stops /__PROJECT__
	if [ "$1" = '/__PROJECT__' ] && ! _want_help "$@"; then
		echo "Entrypoint script for __PROJECT__ Server ${VERSION} started."

		if [ ! -s /etc/__PROJECT__/__PROJECT__.yaml ]; then
		  echo "Building out template for file"
		  /spruce merge /__PROJECT__.yaml /tmp/__PROJECT___spruce.yaml > /etc/__PROJECT__/__PROJECT__.yaml
		fi
	fi

	exec "$@"
}

# If we are sourced from elsewhere, don't perform any further actions
if ! _is_sourced; then
	_main "$@"
fi