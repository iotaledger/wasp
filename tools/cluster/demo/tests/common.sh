set -euo pipefail

err_report() {
    echo "Error on line $1" >&2
}

trap 'err_report $LINENO' ERR

ARGS="$*"

function wasp-cli() {
	(PS4=; set -x; : wasp-cli -w $ARGS "$@")
    command wasp-cli -w $ARGS "$@"
}

rm -f wasp-cli.json owner.json
