package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/droidvirt/droidvirt-ctrl/pkg/apis"
	"github.com/droidvirt/droidvirt-ctrl/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type APIHandler struct {
	namespace string

	kubeConfig *rest.Config
	client     client.Client
	kubeClient kubernetes.Interface
}

func NewAPIHandler(namespace string) (*APIHandler, error) {
	kubeConfig, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	// setup client set
	clientset, err := setupClient(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup kubernetes client: %v", err)
	}

	// setup kubernetes rest client
	kubeClient, err := utils.NewKubeClient()
	if err != nil {
		return nil, fmt.Errorf("failed to setup kubernetes client: %v", err)
	}

	apiHandler := &APIHandler{
		namespace:  namespace,
		client:     clientset,
		kubeClient: kubeClient,
		kubeConfig: kubeConfig,
	}

	return apiHandler, nil
}

func setupClient(config *rest.Config) (client.Client, error) {
	addToSchemeFuncs := []func(s *runtime.Scheme) error{
		apis.AddToScheme,
		corev1.AddToScheme,
	}
	scheme := runtime.NewScheme()
	for _, addToSchemeFunc := range addToSchemeFuncs {
		if err := addToSchemeFunc(scheme); err != nil {
			return nil, err
		}
	}

	clientset, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

type Message struct {
	Message string `json:"message"`
}

func responseJSON(body interface{}, w http.ResponseWriter, statusCode int) {
	jsonResponse, err := json.Marshal(body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(jsonResponse)
}

func responseError(err error, w http.ResponseWriter) {
	responseJSON(Message{err.Error()}, w, http.StatusInternalServerError)
}

func parseBody(r *http.Request, t reflect.Type) (interface{}, error) {
	bodyObj := reflect.New(t).Interface()

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}

	if len(body) != 0 {
		if err := json.Unmarshal(body, &bodyObj); err != nil {
			return nil, err
		}
	}

	return bodyObj, nil
}
