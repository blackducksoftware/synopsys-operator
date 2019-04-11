package actions

import (
	"fmt"
	"sort"
	"strings"

	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/latest"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	bdutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// This file is generated by Buffalo. It offers a basic structure for
// adding, editing and deleting a page. If your model is more
// complex or you need more than the basic implementation you need to
// edit this file.

// Following naming logic is implemented in Buffalo:
// Model: Singular (Blackduck)
// DB Table: Plural (Blackducks)
// Resource: Plural (Blackducks)
// Path: Plural (/blackducks)
// View Template Folder: Plural (/templates/blackducks/)

// BlackducksResource is the resource for the Blackduck model
type BlackducksResource struct {
	buffalo.Resource
	kubeClient      *kubernetes.Clientset
	blackduckClient *blackduckclientset.Clientset
}

// NewBlackduckResource will instantiate the Black Duck Resource
func NewBlackduckResource(kubeConfig *rest.Config) (*BlackducksResource, error) {
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create kube client due to %+v", err)
	}
	hubClient, err := blackduckclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create hub client due to %+v", err)
	}
	return &BlackducksResource{kubeClient: kubeClient, blackduckClient: hubClient}, nil
}

// List gets all Hubs. This function is mapped to the path
// GET /blackducks
func (v BlackducksResource) List(c buffalo.Context) error {
	blackducks, err := util.ListHubs(v.blackduckClient, "")
	if err != nil {
		return c.Error(500, err)
	}
	// Make blackducks available inside the html template
	c.Set("blackducks", blackducks.Items)
	return c.Render(200, r.HTML("blackducks/index.html", "old_application.html"))
}

// Show gets the data for one Blackduck. This function is mapped to
// the path GET /hubs/{hub_id}
func (v BlackducksResource) Show(c buffalo.Context) error {
	blackduck, err := util.GetHub(v.blackduckClient, c.Param("blackduck_id"), c.Param("blackduck_id"))
	if err != nil {
		return c.Error(500, err)
	}
	// Make blackduck available inside the html template
	c.Set("blackduck", blackduck)
	return c.Render(200, r.HTML("blackducks/show.html", "old_application.html"))
}

// New renders the form for creating a new Blackduck.
// This function is mapped to the path GET /blackducks/new
func (v BlackducksResource) New(c buffalo.Context) error {
	blackduckSpec := bdutil.GetHubDefaultValue()
	blackduck := &blackduckapi.Blackduck{}
	blackduck.Spec = *blackduckSpec
	blackduck.Spec.PersistentStorage = true
	blackduck.Spec.PVCStorageClass = ""
	blackduck.Spec.ScanType = "Artifacts"
	blackduck.Spec.PVC = []blackduckapi.PVC{
		{
			Name: "blackduck-postgres",
			Size: "150Gi",
		},
	}

	// Required so that the UI can update the fields
	blackduck.Spec.ExternalPostgres = &blackduckapi.PostgresExternalDBConfig{}
	err := v.common(c, blackduck)
	if err != nil {
		return err
	}
	// Make blackduck available inside the html template
	c.Set("blackduck", blackduck)

	return c.Render(200, r.HTML("blackducks/new.html", "old_application.html"))
}

func (v BlackducksResource) common(c buffalo.Context, bd *blackduckapi.Blackduck) error {
	var storageList map[string]string
	storageList = make(map[string]string)
	storageClasses, err := util.ListStorageClasses(v.kubeClient)
	if err != nil {
		c.Error(404, fmt.Errorf("\"message\": \"Failed to List the storage class due to %+v\"", err))
	}
	for _, storageClass := range storageClasses.Items {
		storageList[fmt.Sprintf("%s (%s)", storageClass.GetName(), storageClass.Provisioner)] = storageClass.GetName()
	}
	storageList[fmt.Sprintf("%s (%s)", "None", "Disable dynamic provisioner")] = ""
	bd.View.StorageClasses = storageList

	// Blackduck instances
	keys := make(map[string]string)
	hubs, _ := util.ListHubs(v.blackduckClient, "")
	for _, v := range hubs.Items {
		if strings.EqualFold(v.Status.State, "running") {
			keys[v.Name] = v.Name
		}
	}
	keys["None"] = ""
	bd.View.Clones = keys

	blackducks, _ := util.ListHubs(v.blackduckClient, "")
	certificateNames := []string{"default", "manual"}
	for _, hub := range blackducks.Items {
		if strings.EqualFold(hub.Spec.CertificateName, "manual") {
			certificateNames = append(certificateNames, hub.Spec.Namespace)
		}
	}
	bd.View.CertificateNames = certificateNames

	env, _ := blackduck.GetHubKnobs()

	environs := []string{}
	for key, value := range env {
		if !strings.EqualFold(value, "") {
			environs = append(environs, fmt.Sprintf("%s:%s", key, value))
		}
	}

	if len(bd.Spec.Environs) > 0 {
		bd.View.Environs = bd.Spec.Environs
	} else {
		bd.View.Environs = environs
	}

	kubeconfig, err := protoform.GetKubeConfig()
	if err != nil {
		return nil
	}

	bd.View.SupportedVersions = apps.NewApp(nil, kubeconfig).Blackduck().Versions()
	sort.Sort(sort.Reverse(sort.StringSlice(bd.View.SupportedVersions)))

	return nil
}

func (v BlackducksResource) redirect(c buffalo.Context, blackduck *blackduckapi.Blackduck, err error) error {
	if err != nil {
		c.Flash().Add("warning", err.Error())
		// Make blackduck available inside the html template
		err = v.common(c, blackduck)
		if err != nil {
			log.Error(err)
			return err
		}
		log.Debugf("edit hub in create: %+v", blackduck)

		c.Set("blackduck", blackduck)

		// validateErrs := validate.NewErrors()
		// log.Infof("Error: %s", err.Error())
		// validateErrs.Add("error", err.Error())
		// log.Infof("validateErrs: %+v", validateErrs)
		// validateErrs.Errors = map[string][]string{"error": []string{errors.WithStack(err).Error()}}
		// c.Set("errors", err.Error())
		return c.Render(422, r.HTML("blackducks/new.html", "old_application.html"))
	}
	return nil
}

// Create adds a Blackduck to the DB. This function is mapped to the
// path POST /blackducks
func (v BlackducksResource) Create(c buffalo.Context) error {
	// Allocate an empty Blackduck
	blackduck := &blackduckapi.Blackduck{}

	// Bind blackduck to the html form elements
	if err := c.Bind(blackduck); err != nil {
		log.Errorf("unable to bind blackduck %+v because %+v", c, err)
		return errors.WithStack(err)
	}

	log.Infof("create blackduck: %+v", blackduck)

	_, err := util.GetHub(v.blackduckClient, blackduck.Spec.Namespace, blackduck.Spec.Namespace)

	if err == nil {
		return v.redirect(c, blackduck, fmt.Errorf("blackduck %s already exist", blackduck.Spec.Namespace))
	}

	_, err = util.GetNamespace(v.kubeClient, blackduck.Spec.Namespace)

	if err == nil {
		return v.redirect(c, blackduck, fmt.Errorf("namespace %s already exist", blackduck.Spec.Namespace))
	}

	ns, err := util.CreateNamespace(v.kubeClient, blackduck.Spec.Namespace)
	if err != nil {
		return v.redirect(c, blackduck, err)
	}
	log.Infof("created namespace for %s is %+v", blackduck.Spec.Namespace, ns)

	if !blackduck.Spec.PersistentStorage {
		blackduck.Spec.PVC = nil
	} else {
		// Remove postgres volume if we use an external db
		if *blackduck.Spec.ExternalPostgres != (blackduckapi.PostgresExternalDBConfig{}) {
			blackduck.Spec.PVC = nil
		}
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-authentication"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-cfssl"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-registration"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-solr"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-webapp"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-logstash"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-zookeeper-data"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-zookeeper-datalog"})
	}

	// Change back to nil if the configuration is empty
	if *blackduck.Spec.ExternalPostgres == (blackduckapi.PostgresExternalDBConfig{}) {
		log.Info("External Database configuration is empty")
		blackduck.Spec.ExternalPostgres = nil
	}
	_, err = util.CreateHub(v.blackduckClient, blackduck.Spec.Namespace, &blackduckapi.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: blackduck.Spec.Namespace}, Spec: blackduck.Spec})

	if err != nil {
		return v.redirect(c, blackduck, err)
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "Black Duck was created successfully")

	blackducks, _ := util.ListHubs(v.blackduckClient, "")
	c.Set("blackducks", blackducks.Items)
	// and redirect to the blackducks index page
	return c.Redirect(302, "/blackducks/%s", blackduck.Spec.Namespace)
}

// Edit renders a edit form for a Blackduck. This function is
// mapped to the path GET /blackducks/{blackduck_id}/edit
func (v BlackducksResource) Edit(c buffalo.Context) error {
	blackduck, err := util.GetHub(v.blackduckClient, c.Param("blackduck_id"), c.Param("blackduck_id"))
	if err != nil {
		return c.Error(404, err)
	}
	if blackduck.Spec.ExternalPostgres == nil {
		blackduck.Spec.ExternalPostgres = &blackduckapi.PostgresExternalDBConfig{}
	}
	blackduck.Spec.PVC = []blackduckapi.PVC{
		{
			Name: "blackduck-postgres",
			Size: "150Gi",
		},
	}
	err = v.common(c, blackduck)
	if err != nil {
		return c.Error(500, err)
	}
	return c.Render(200, r.Auto(c, blackduck))
}

// Update changes a Blackduck in the DB. This function is mapped to
// the path PUT /blackducks/{blackduck_id}
func (v BlackducksResource) Update(c buffalo.Context) error {
	// Allocate an empty Blackduck
	blackduck := &blackduckapi.Blackduck{}

	// Bind blackduck to the html form elements
	if err := c.Bind(blackduck); err != nil {
		log.Errorf("unable to bind blackduck %+v because %+v", c, err)
		return errors.WithStack(err)
	}

	if !blackduck.Spec.PersistentStorage {
		blackduck.Spec.PVC = nil
	} else {
		// Remove postgres volume if we use an external db
		if *blackduck.Spec.ExternalPostgres != (blackduckapi.PostgresExternalDBConfig{}) {
			blackduck.Spec.PVC = nil
		}
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-authentication"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-cfssl"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-registration"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-solr"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-webapp"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-logstash"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-zookeeper-data"})
		blackduck.Spec.PVC = append(blackduck.Spec.PVC, blackduckapi.PVC{Name: "blackduck-zookeeper-datalog"})
	}

	// Change back to nil if the configuration is empty
	if *blackduck.Spec.ExternalPostgres == (blackduckapi.PostgresExternalDBConfig{}) {
		log.Info("External Database configuration is empty")
		blackduck.Spec.ExternalPostgres = nil
	}

	latestBlackduck, err := util.GetHub(v.blackduckClient, blackduck.Spec.Namespace, blackduck.Spec.Namespace)
	if err != nil {
		log.Errorf("unable to get %s blackduck instance because %+v", blackduck.Spec.Namespace, err)
		return v.redirect(c, blackduck, err)
	}

	latestBlackduck.Spec = blackduck.Spec
	_, err = util.UpdateBlackduck(v.blackduckClient, blackduck.Spec.Namespace, latestBlackduck)

	if err != nil {
		log.Errorf("unable to update %s blackduck instance because %+v", blackduck.Spec.Namespace, err)
		return v.redirect(c, blackduck, err)
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "Black Duck was updated successfully")

	blackducks, _ := util.ListHubs(v.blackduckClient, "")
	c.Set("blackducks", blackducks.Items)
	// and redirect to the blackducks index page
	return c.Redirect(302, "/blackducks/%s", blackduck.Spec.Namespace)

	// 	// Render again the edit.html template that the user can
	// 	// correct the input.
	// 	return c.Render(422, r.Auto(c, blackduck))
	// }

	// // If there are no errors set a success message
	// c.Flash().Add("success", "Blackduck was updated successfully")

	// and redirect to the blackducks index page
}

// Destroy deletes a Blackduck from the DB. This function is mapped
// to the path DELETE /blackducks/{blackduck_id}
func (v BlackducksResource) Destroy(c buffalo.Context) error {

	log.Infof("delete blackduck request %v", c.Param("blackduck"))

	_, err := util.GetHub(v.blackduckClient, c.Param("blackduck_id"), c.Param("blackduck_id"))
	// To find the Blackduck the parameter blackduck_id is used.
	if err != nil {
		return c.Error(404, err)
	}

	// This is on the event loop.
	err = v.blackduckClient.SynopsysV1().Blackducks(c.Param("blackduck_id")).Delete(c.Param("blackduck_id"), &metav1.DeleteOptions{})

	// To find the Blackduck the parameter blackduck_id is used.
	if err != nil {
		return c.Error(404, err)
	}

	// If there are no errors set a flash message
	c.Flash().Add("success", "Blackduck was deleted successfully")

	// blackducks, _ := util.ListHubs(v.blackduckClient, "")
	// c.Set("hubs", blackducks.Items)

	// Redirect to the blackducks index page
	return c.Redirect(302, "/blackducks")
}
