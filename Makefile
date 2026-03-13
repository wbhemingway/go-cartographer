sinclude .env
export

FULL_IMAGE_PATH=$(CARTOGRAPHER_REGION)-docker.pkg.dev/$(CARTOGRAPHER_PROJECT_ID)/$(CARTOGRAPHER_REPO)/$(CARTOGRAPHER_IMAGE_NAME):latest

.PHONY: build push deploy all

build:
	docker build -f server.Dockerfile -t $(CARTOGRAPHER_IMAGE_NAME) .

push: build
	docker tag $(CARTOGRAPHER_IMAGE_NAME) $(FULL_IMAGE_PATH)
	docker push $(FULL_IMAGE_PATH)

deploy: push
	gcloud run deploy cartographer-api \
		--image $(FULL_IMAGE_PATH) \
		--region $(CARTOGRAPHER_REGION) \
		--set-secrets="CARTOGRAPHER_API_KEY=CARTOGRAPHER_API_KEY:latest" \
		--set-env-vars="CARTOGRAPHER_BUCKET_NAME=$(CARTOGRAPHER_BUCKET_NAME)" \
		--allow-unauthenticated \
		--max-instances=2

all: deploy
