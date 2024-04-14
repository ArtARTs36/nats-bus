test:
	docker-compose up -d nats1 nats2 nats3
	sleep 5
	docker-compose run nats-bus-functests
	docker-compose down

lint:
	golangci-lint run --fix
