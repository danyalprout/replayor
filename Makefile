build:
	docker-compose build

run:
	docker-compose down
	docker-compose build
	rm -rf ./geth-data
	cp -R ./geth-data-archive ./geth-data
	docker-compose up	-d
	docker-compose logs -f replayor

init:
	mkdir secret
	openssl rand -hex 32 > secret/jwt