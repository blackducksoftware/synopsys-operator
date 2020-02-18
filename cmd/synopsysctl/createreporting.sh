# kubectl create secret tls -n reporting polaris-reporting-ingress-tls --cert="$TLSOUT_CRT" --key="$TLSOUT_KEY"

./synopsysctl create polaris-reporting -v debug \
--namespace="reporting" \
--gcp-service-account-path "/Users/manikan/work/src/github.com/blackducksoftware/polaris-contrib/snps-swip-staging-308eb0be99bd.json" \
--fqdn="onprem-dev.dev.polaris.synopsys.com" \
--storage-class="standard" \
--reportstorage-size="1Gi" \
--eventstore-size="1Gi" \
--smtp-host="smtp.sendgrid.net " \
--smtp-port="2525" \
--smtp-username="apikey" \
--smtp-password="$SMTP_PASSWORD" \
--smtp-sender-email="noreply@synopsys.com" \
--postgres-host="postgres" \
--postgres-username="postgres" \
--postgres-password="admin" \
--enable-postgres-container=true \
--postgres-size="1Gi" \
--kubeconfig="/Users/manikan/.kube/config"

# --chart-location-path="/Users/hammer/go/src/github.com/blackducksoftware/polaris-helmchart-reporting"



# --version "" \
# --postgres-ssl-mode "" \
# --postgres-port "" \
# --smtp-tls-mode "disable" \
# --smtp-trusted-hosts "" \
# --insecure-skip-smtp-tls-verify true \