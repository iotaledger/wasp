all: build

build:
	make -C .. build install
	cd .. && docker build -t isc .

run-cluster:
	docker compose up

run-spam:
	for i in 0 1 2 3 4 5 6; do ./spam.sh $i & done

stop-spam:
	jobs -p | xargs -I{} kill -- -{}
