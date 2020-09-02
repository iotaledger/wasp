set -euo pipefail

err_report() {
    echo "Error on line $1" >&2
}

trap 'err_report $LINENO' ERR

ARGS="$*"

function wwallet() {
	(PS4=; set -x; : wwallet -w $ARGS "$@")
    command wwallet -w $ARGS "$@"
}
