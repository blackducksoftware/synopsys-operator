package webhook

import (
	"encoding/json"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	v1beta12 "k8s.io/api/authentication/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"reflect"
	"strings"
)

func (ow *OperatorWebhook) serveCustomResource(w http.ResponseWriter, r *http.Request) {
	ow.serve(w, r, ow.blackduckCustomResource)
}

func (ow *OperatorWebhook) blackduckCustomResource(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	bd := v1.Blackduck{}
	err := json.Unmarshal(ar.Request.Object.Raw, &bd)
	if err != nil {
		logrus.Error(err)
		return ow.returnError(err.Error())
	}

	// Get TokenReview for current user
	tok, err := ow.kubeClient.AuthenticationV1beta1().TokenReviews().Create(&v1beta12.TokenReview{Spec: v1beta12.TokenReviewSpec{Token: ow.kubeConfig.BearerToken}})
	if err != nil {
		return ow.returnError(err.Error())
	}

	// Log the request
	logrus.Infof("Resource: %s, Operation: %s\nUsername: %s\n\n", ar.Request.Name, ar.Request.Operation, ar.Request.UserInfo.Username)

	reviewResponse := v1beta1.AdmissionResponse{}

	// Only if the resource is being updated and that request was made by a different user
	if ar.Request.Operation == v1beta1.Update && !strings.EqualFold(tok.Status.User.Username, ar.Request.UserInfo.Username) {
		current, err := ow.blackduckClient.SynopsysV1().Blackducks(corev1.NamespaceDefault).Get(bd.Name, metav1.GetOptions{})
		if err != nil {
			logrus.Error(err)
			return ow.returnError(err.Error())
		}

		if !reflect.DeepEqual(current.Status, bd.Status) {
			logrus.Error(err)
			return ow.returnError("Status cannot be modified")
		}

		if strings.Compare(current.Spec.Namespace, bd.Spec.Namespace) != 0 {
			return ow.returnError("Namespace cannot be modified")
		}
	}

	reviewResponse.Allowed = true
	return &reviewResponse
}
