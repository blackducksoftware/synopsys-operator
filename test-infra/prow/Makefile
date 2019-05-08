# Modified from: https://github.com/openshift/release/blob/master/Makefile and https://github.com/Azure/aks-engine/blob/master/.prowci/Makefile

prow: prow-crd prow-services prow-rbac
.PHONY: prow

prow-crd:
	kubectl apply -f prow_crd.yaml
	kubectl apply -f prow_misc.yaml
.PHONY: prow-crd

prow-config:
	kubectl create cm config --from-file=config.yaml=config.yaml
	kubectl create cm plugins --from-file=plugins.yaml=plugins.yaml
.PHONY: prow-config

prow-config-update:
	kubectl create cm config --from-file=config.yaml=config.yaml -o yaml --dry-run | kubectl replace -f -
	kubectl create cm plugins --from-file=plugins.yaml=plugins.yaml -o yaml --dry-run | kubectl replace -f -
.PHONY: prow-config-update

prow-secrets:
	# hmac is used for encrypting Github webhook payloads.
	kubectl create secret generic hmac-token --from-file=hmac
	# oauth is used for merging PRs, adding/removing labels and comments.
	kubectl create secret generic oauth-token --from-file=oauth
.PHONY: prow-secrets

prow-services:
	kubectl apply -f deck.yaml
	kubectl apply -f hook.yaml
	kubectl apply -f horologium.yaml
	kubectl apply -f plank.yaml
	kubectl apply -f sinker.yaml
	kubectl apply -f tide.yaml
	kubectl apply -f statusreconciler.yaml
	kubectl apply -f cherrypick.yaml
	kubectl apply -f needs_rebase.yaml
.PHONY: prow-services

prow-rbac:
	kubectl apply -f deck_rbac.yaml
	kubectl apply -f hook_rbac.yaml
	kubectl apply -f horologium_rbac.yaml
	kubectl apply -f plank_rbac.yaml
	kubectl apply -f sinker_rbac.yaml
	kubectl apply -f tide_rbac.yaml
	kubectl apply -f statusreconciler_rbac.yaml
.PHONY: prow-rbac

