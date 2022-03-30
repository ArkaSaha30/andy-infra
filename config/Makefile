# AWS variables.
AWS_REGION ?= us-east-1

update-config: kubectl --context=prow-service-trusted create configmap config --from-file=config.yaml=config.yaml --dry-run=client -o yaml | kubectl -n prow replace configmap config -f -

update-plugins: kubectl --context=prow-service-trusted create configmap plugins --from-file=plugins.yaml=plugins.yaml --dry-run=client -o yaml | kubectl -n prow replace configmap plugins -f -

deploy: kubectl --context=prow-service-trusted apply -f ./prow/cluster/

.PHONY: deploy update-config update-plugins
