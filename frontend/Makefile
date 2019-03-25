.PHONY: frontend_img frontend_deploy fuzzer-client

GIT_SHA = $(shell bash ../scripts/get_git_sha.sh)

fuzzer-client:
	docker build -f ../dockerfiles/client.dockerfile -t gcr.io/athena-fuzzer/athena:$(GIT_SHA) ..
	docker push gcr.io/athena-fuzzer/athena:$(GIT_SHA)
