.PHONY: frontend

GIT_SHA = $(shell git log | head -n 1 | cut -f 2 -d ' ' | head -c 10)

frontend:
	cp ~/.kube/config  kubeconfig.json
	docker build -t frontend:$(GIT_SHA) .
