build:
	docker build -t panjiang/certificate-sentry .
	docker tag panjiang/certificate-sentry panjiang/certificate-sentry:latest
	docker push panjiang/certificate-sentry:latest