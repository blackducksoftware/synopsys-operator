package webhook

import (
	"encoding/json"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
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

	reviewResponse := v1beta1.AdmissionResponse{}
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

	reviewResponse.Allowed = true
	return &reviewResponse
}
