package webhook

import (
	"encoding/json"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"time"

	"k8s.io/api/admission/v1beta1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// OperatorWebhook is used to create the admission webhook
type OperatorWebhook struct {
	kubeConfig      *rest.Config
	kubeClient      *kubernetes.Clientset
	blackduckClient *blackduckclientset.Clientset
}

// NewOperatorWebhook will return an OperatorWebhook
func NewOperatorWebhook(kubeConfig *rest.Config) *OperatorWebhook {
	kubeclient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}
	blackduckClient, err := blackduckclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}
	return &OperatorWebhook{
		kubeConfig:      kubeConfig,
		kubeClient:      kubeclient,
		blackduckClient: blackduckClient,
	}
}

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

func init() {
	addToScheme(scheme)
}

func addToScheme(scheme *runtime.Scheme) {
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(admissionv1beta1.AddToScheme(scheme))
	utilruntime.Must(admissionregistrationv1beta1.AddToScheme(scheme))
}

func (ow *OperatorWebhook) returnError(message string) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Message: message,
		},
	}
}

type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

func (ow *OperatorWebhook) serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logrus.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := v1beta1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := v1beta1.AdmissionReview{}

	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		logrus.Error(err)
		responseAdmissionReview.Response = ow.returnError(err.Error())
	} else {
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		logrus.Error(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		logrus.Error(err)
	}
}

// Start will start the web server
func (ow *OperatorWebhook) Start() {
	server := &http.Server{
		Addr:         ":443",
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	http.HandleFunc("/hook/custom-resource/blackduck", ow.serveCustomResource)

	err := server.ListenAndServeTLS("/opt/synopsys-operator/tls/cert.crt", "/opt/synopsys-operator/tls/cert.key")
	if err != nil {
		logrus.Error(err)
	}
}
