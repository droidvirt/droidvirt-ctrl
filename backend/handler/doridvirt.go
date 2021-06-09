package handler

import (
	"context"
	"net/http"
	"reflect"

	"github.com/droidvirt/droidvirt-ctrl/backend/types"
	dvv1alpha1 "github.com/droidvirt/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *APIHandler) ListDroidVirt(w http.ResponseWriter, r *http.Request) {
	var body types.QueryDroidVirtReq
	rawBody, err := parseBody(r, reflect.TypeOf(body))
	if err != nil {
		responseError(err, w)
		return
	}
	body = rawBody.(types.QueryDroidVirtReq)

	selectors := []fields.Selector{}
	if body.UID != "" {
		selectors = append(selectors, fields.OneTermEqualSelector("metadata.uid", body.UID))
	}
	listOpts := &client.ListOptions{
		Namespace:     h.namespace,
		FieldSelector: fields.AndSelectors(selectors...),
	}

	virtList := &dvv1alpha1.DroidVirtList{}
	err = h.client.List(context.TODO(), listOpts, virtList)
	if err != nil {
		responseError(err, w)
		return
	}


}
