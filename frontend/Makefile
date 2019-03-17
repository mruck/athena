.PHONY: frontend_img frontend_deploy

GIT_SHA = $(shell git log | head -n 1 | cut -f 2 -d ' ' | head -c 10)

frontend_img:
	cp ~/.kube/config  k8s/kubeconfig.yaml
	docker build -t gcr.io/athena-fuzzer/frontend:$(GIT_SHA) .
	docker push gcr.io/athena-fuzzer/frontend:$(GIT_SHA)

frontend_deploy: frontend_img
	jq '.spec.template.spec.containers[0].image = "gcr.io/athena-fuzzer/frontend:'$(GIT_SHA)'"' k8s/frontend.daemonset.template.json > /tmp/frontend.daemonset.json
	kubectl apply -f /tmp/frontend.daemonset.json
