# AWS variables.
AWS_REGION ?= us-east-1

set-context:
	kubectl config use-context prow-service-trusted

update-config: set-context
		kubectl create configmap config --from-file=config.yaml=config.yaml --dry-run=client -o yaml | kubectl -n prow replace configmap config -f -

update-plugins: set-context
		kubectl create configmap plugins --from-file=plugins.yaml=plugins.yaml --dry-run=client -o yaml | kubectl -n prow replace configmap plugins -f -

deploy: set-context
	kubectl apply -f ./prow/cluster/

.PHONY: deploy set-context update-config update-plugins
