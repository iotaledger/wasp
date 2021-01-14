set -euo pipefail

err_report() {
    echo "Error on line $1" >&2
}

trap 'err_report $LINENO' ERR

function wasp-cli() {
	(PS4=; set -x; : wasp-cli -w -d "$@")
    command wasp-cli -w -d "$@"
}

rm -f wasp-cli.json owner.json

if [ "$*" = "-u" ]; then
	wasp-cli set utxodb true
fi
