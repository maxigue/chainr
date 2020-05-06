ALL_SERVICES=gate $(GO_SERVICES)
GO_SERVICES=sched work notif

GO_TARGETS=all build test fmt check-fmt docker-build

.PHONY: $(GO_TARGETS) $(GO_SERVICES) run deploy clean

$(GO_TARGETS):
	for s in $(GO_SERVICES); do \
	  $(MAKE) $@ -C $$s || exit 1; \
	done

$(GO_SERVICES):
	$(MAKE) -C $@

run:
	@echo "To run a service, use \"make run -C <service>\""

deploy:
	./scripts/deploy.sh; \
	for s in $(ALL_SERVICES); do \
	  $(MAKE) $@ -C $$s || exit 1; \
	done

clean:
	for s in $(ALL_SERVICES); do \
	  $(MAKE) $@ -C $$s || exit 1; \
	done
