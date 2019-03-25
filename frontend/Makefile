.PHONY: frontend_img frontend_deploy fuzzer-client

GIT_SHA = $(shell bash ../scripts/get_git_sha.sh)

fuzzer-client:
	docker build -f ../dockerfiles/client.dockerfile -t gcr.io/athena-fuzzer/athena:$(GIT_SHA) ..
	docker push gcr.io/athena-fuzzer/athena:$(GIT_SHA)

frontend_img:
	cp ~/.kube/config  k8s/kubeconfig.yaml
	docker build -t gcr.io/athena-fuzzer/frontend:$(GIT_SHA) .
	docker push gcr.io/athena-fuzzer/frontend:$(GIT_SHA)

frontend_deploy: frontend_img fuzzer-client
	jq '.spec.template.spec.containers[0].image = "gcr.io/athena-fuzzer/frontend:'$(GIT_SHA)'"' k8s/frontend.daemonset.template.json | jq '.spec.template.spec.containers[0].env[0].value = "gcr.io/athena-fuzzer/athena:'$(GIT_SHA)'"' > /tmp/frontend.daemonset.json
	kubectl apply -f /tmp/frontend.daemonset.json
