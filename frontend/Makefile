.PHONY: frontend

GIT_SHA = $(shell git log | head -n 1 | cut -f 2 -d ' ' | head -c 10)

frontend_img:
	cp ~/.kube/config  k8s/kubeconfig.yaml
	docker build -t frontend:$(GIT_SHA) .
