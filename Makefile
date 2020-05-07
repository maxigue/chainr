ALL_SERVICES=gate $(GO_SERVICES)
GO_SERVICES=sched work notif

.PHONY: all build sched work notif ui run test fmt check-fmt docker-build deploy clean

all: build

build: sched work notif ui

deps:
	npm install -C ui

sched:
	make build -C sched

work:
	make build -C work

notif:
	make build -C notif

ui:
	npm run build -C ui

run:
	@echo "To run a service, use \"make run -C <service>\""

test:
	make test -C sched
	make test -C work
	make test -C notif
	npm run test:unit -C ui

fmt:
	make fmt -C sched
	make fmt -C work
	make fmt -C notif
	npm run lint -C ui

check-fmt:
	make check-fmt -C sched
	make check-fmt -C work
	make check-fmt -C notif
	npm run lint-no-fix -C ui

docker-build:
	make docker-build -C sched
	make docker-build -C work
	make docker-build -C notif
	npm run docker-build -C ui

docker-push:
	make docker-push -C sched
	make docker-push -C work
	make docker-push -C notif
	npm run docker-push -C ui

deploy:
	./scripts/deploy.sh
	make deploy -C gate
	make deploy -C sched
	make deploy -C work
	make deploy -C notif
	npm run deploy -C ui

clean:
	make clean -C gate
	make clean -C sched
	make clean -C work
	make clean -C notif
	npm run clean -C ui
