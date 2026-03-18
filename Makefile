sinclude .env
export

API_IMAGE_PATH=$(CARTOGRAPHER_REGION)-docker.pkg.dev/$(CARTOGRAPHER_PROJECT_ID)/$(CARTOGRAPHER_REPO)/cartographer-api:latest
WORKER_IMAGE_PATH=$(CARTOGRAPHER_REGION)-docker.pkg.dev/$(CARTOGRAPHER_PROJECT_ID)/$(CARTOGRAPHER_REPO)/cartographer-worker:latest

.PHONY: build-api push-api deploy-api build-worker push-worker deploy-worker build-all push-all deploy-all all

# ==========================================
# API TARGETS
# ==========================================
build-api:
	docker build -f api.Dockerfile -t cartographer-api .

push-api: build-api
	docker tag cartographer-api $(API_IMAGE_PATH)
	docker push $(API_IMAGE_PATH)

deploy-api: push-api
	gcloud run deploy cartographer-api \
		--image $(API_IMAGE_PATH) \
		--region $(CARTOGRAPHER_REGION) \
		--set-env-vars="CARTOGRAPHER_BUCKET_NAME=$(CARTOGRAPHER_BUCKET_NAME)" \
		--allow-unauthenticated \
		--max-instances=2

# ==========================================
# WORKER TARGETS
# ==========================================
build-worker:
	docker build -f worker.Dockerfile -t cartographer-worker .

push-worker: build-worker
	docker tag cartographer-worker $(WORKER_IMAGE_PATH)
	docker push $(WORKER_IMAGE_PATH)

deploy-worker: push-worker
	gcloud run deploy cartographer-worker \
		--image $(WORKER_IMAGE_PATH) \
		--region $(CARTOGRAPHER_REGION) \
		--set-env-vars="CARTOGRAPHER_BUCKET_NAME=$(CARTOGRAPHER_BUCKET_NAME)" \
		--no-allow-unauthenticated \
		--min-instances=1 \
		--no-cpu-throttling \
		--max-instances=2

# ==========================================
# BATCH TARGETS
# ==========================================
build-all: build-api build-worker

push-all: push-api push-worker

deploy-all: deploy-api deploy-worker

all: deploy-all