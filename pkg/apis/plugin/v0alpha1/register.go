package v0alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/grafana/grafana/pkg/apimachinery/utils"
)

const (
	GROUP      = "plugin.grafana.app"
	VERSION    = "v0alpha1"
	APIVERSION = GROUP + "/" + VERSION
	KIND       = "Plugin"
)

var PluginResourceInfo = utils.NewResourceInfo(GROUP, VERSION,
	"plugins", "plugin", KIND,
	func() runtime.Object { return &PluginResource{} },
	func() runtime.Object { return &PluginResourceList{} },
	utils.TableColumns{
		Definition: []metav1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name"},
			{Name: "Id", Type: "string", Format: "id"},
			{Name: "Version", Type: "string", Format: "string"},
		},
		Reader: func(obj any) ([]interface{}, error) {
			m, ok := obj.(*PluginResource)
			if ok {
				return []interface{}{
					m.Name,
					m.Spec.ID,
					m.Spec.Version,
				}, nil
			}
			return nil, fmt.Errorf("expected Plugin but got %T", obj)
		},
	}, // default table converter
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: GROUP, Version: VERSION}

	// SchemeBuilder is used by standard codegen
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)
