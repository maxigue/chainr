SERVICES=sched work notif
TARGETS=all build test fmt check-fmt docker-build deploy clean

.PHONY: $(TARGETS) $(SERVICES) run

$(TARGETS):
	if [ $@ == "deploy" ]; then \
		./scripts/deploy.sh; \
	fi; \
	for s in $(SERVICES); do \
	  $(MAKE) $@ -C $$s || exit 1; \
	done

$(SERVICES):
	$(MAKE) -C $@

run:
	@echo "To run a service, use \"make run -C <service>\""
