ALL_SERVICES=gate $(GO_SERVICES)
GO_SERVICES=sched work notif recycle

.PHONY: all build sched work notif recycle ui run test fmt check-fmt docker-build deploy clean

all: build

build: sched work notif recycle ui

sched:
	make -C sched

work:
	make -C work

notif:
	make -C notif

recycle:
	make -C recycle

ui:
	npm run build -C ui

run:
	@echo "To run a service, use \"make run -C <service>\""

test:
	make test -C sched
	make test -C work
	make test -C notif
	make test -C recycle
	npm run test:unit -C ui

fmt:
	make fmt -C sched
	make fmt -C work
	make fmt -C notif
	make fmt -C recycle
	npm run lint -C ui

check-fmt:
	make check-fmt -C sched
	make check-fmt -C work
	make check-fmt -C notif
	make check-fmt -C recycle
	npm run lint:no-fix -C ui

docker-build:
	make docker-build -C sched
	make docker-build -C work
	make docker-build -C notif
	make docker-build -C recycle
	npm run docker:build -C ui

docker-push:
	make docker-push -C sched
	make docker-push -C work
	make docker-push -C notif
	make docker-push -C recycle
	npm run docker:push -C ui

deploy:
	./scripts/deploy.sh
	make deploy -C gate
	make deploy -C sched
	make deploy -C work
	make deploy -C notif
	make deploy -C recycle
	npm run deploy -C ui

chaos:
	./scripts/chaos.sh

clean:
	make clean -C gate
	make clean -C sched
	make clean -C work
	make clean -C notif
	make clean -C recycle
	npm run clean -C ui
