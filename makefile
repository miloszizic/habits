SHELL := /bin/bash

run:
	go run ./cmd/main.go

# ======================================================================

VERSION := 1.0

all: service

service:
	docker build \
		-f infra/docker/dockerfile \
		-t service-habits:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ========================================================================

KIND_CLUSTER := starter-cluster

kind-create:
	kind create cluster \
		--name $(KIND_CLUSTER) \
		--image kindest/node-arm64:v1.21.1@sha256:76ec5e9969179acac692ce1ae308cdb8bdaf70c8ecb80a376f7b3a01f58cd206 \
		--config infra/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=service-system

tidy:
	go mod tidy

kind-up:
	docker start $(KIND_CLUSTER)-control-plane

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-status-service:
	kubectl get pods -o wide --watch

kind-load:
	kind load docker-image service-habits:$(VERSION) --name $(KIND_CLUSTER)

kind-apply:
	kustomize build infra/k8s/kind/service-habits | kubectl apply -f -

kind-logs:
	kubectl logs -l app=service-habits --all-containers=true -f --tail=100

kind-restart:
	kubectl rollout restart deployment service-habits

kind-describe:
	kubectl describe nodes
	kubectl describe svc
	kubectl describe pod -l app=service-habits

kind-update: all kind-load kind-restart

kind-update-apply: all kind-load kind-apply

