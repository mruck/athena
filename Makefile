GIT_SHA ?= $(shell bash scripts/get_git_sha.sh)
VENV_LOCATION ?= $(shell pwd)/venv

MITM_TARGET ?= 3000

.PHONY: postgres-start postgres-stop redis-start redis-stop python-test venv debug-deployment frontend_deploy frontend_img

rails:
	GIT_SHA=$(GIT_SHA) $(MAKE) -C ../rails-fork rails

# Bump images in debug deployment
debug-deployment: fuzzer-client discourse-server
	jq '.spec.template.spec.containers[1].image = "gcr.io/athena-fuzzer/discourse:'$(GIT_SHA)'"' pods/debug.deployment.json | jq '.spec.template.spec.containers[2].image = "gcr.io/athena-fuzzer/athena:'$(GIT_SHA)'"' > /tmp/deployments/$(GIT_SHA)
	kubectl apply -f /tmp/deployments/$(GIT_SHA)

postgres-stop:
	-docker rm -f my-postgres

postgres-start: postgres-stop
	docker run --name my-postgres -e POSTGRES_USER="root" -d postgres

venv: pip-reqs.txt
	-rm -rf $(VENV_LOCATION)
	virtualenv -p python3 $(VENV_LOCATION)
	$(VENV_LOCATION)/bin/pip install -r pip-reqs.txt
	ln -s $(shell pwd)/fuzzer $(VENV_LOCATION)/lib/python3.7/site-packages/fuzzer
	printf 'Please run the following:\nsource $(VENV_LOCATION)/bin/activate\n'

fuzzer-db:
	docker build -f dockerfiles/db.dockerfile -t fuzzer-db:$(GIT_SHA) .

run-db:
	docker run -it --volumes-from my-postgres --entrypoint=bash fuzzer-db

fuzzer-client:
	docker build -f dockerfiles/client.dockerfile -t gcr.io/athena-fuzzer/athena:$(GIT_SHA) .
	docker push gcr.io/athena-fuzzer/athena:$(GIT_SHA)

frontend_img:
	docker build -t gcr.io/athena-fuzzer/frontend:$(GIT_SHA) frontend
	docker push gcr.io/athena-fuzzer/frontend:$(GIT_SHA)

frontend_deploy: frontend_img fuzzer-client
	jq '.spec.template.spec.containers[0].image = "gcr.io/athena-fuzzer/frontend:'$(GIT_SHA)'"' frontend/k8s/frontend.daemonset.template.json | jq '.spec.template.spec.containers[0].env[0].value = "gcr.io/athena-fuzzer/athena:'$(GIT_SHA)'"' > /tmp/frontend.daemonset.json
	kubectl apply -f /tmp/frontend.daemonset.json

discourse-server:
	docker build -t gcr.io/athena-fuzzer/discourse:$(GIT_SHA) -f ../discourse-fork/Dockerfile ..
	docker push gcr.io/athena-fuzzer/discourse:$(GIT_SHA)

medium:
	docker build -t medium:$(GIT_SHA) -f ../dante-stories-fork/Dockerfile ..

# To get the client to talk to the proxy, make sure to set the following environment
# variables in the client container:
# export PORT=5443
# export TARGET_HOSTNAME=host.docker.internal
mitm:
	mitmweb --web-iface 0.0.0.0 -v -k -p 5443 --mode reverse:http://127.0.0.1:$(MITM_TARGET)

all: postgres-start fuzzer-db fuzzer-client discourse-server venv

images: fuzzer-db fuzzer-client discourse-server

python-test:
	python ./tests/fuzz_state_test.py
	python ./tests/route_matching_test.py

# Remove lingering containers except for postgres
clean:
	docker ps -a | tail -n+2 | grep -v "postgres" | cut -d ' ' -f 1 | xargs docker rm -f
	rm -rf venv
