package plugin

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/kube-openapi/pkg/common"

	"github.com/grafana/grafana/pkg/apimachinery/identity"
	plugin "github.com/grafana/grafana/pkg/apis/plugin/v0alpha1"
	"github.com/grafana/grafana/pkg/services/apiserver/builder"
	"github.com/grafana/grafana/pkg/services/pluginsintegration/pluginstore"
)

var _ builder.APIGroupBuilder = (*APIBuilder)(nil)

type APIBuilder struct {
	registerer  prometheus.Registerer
	pluginStore pluginstore.Store
}

func RegisterAPIService(apiRegistrar builder.APIRegistrar, pluginStore pluginstore.Store,
	registerer prometheus.Registerer) *APIBuilder {
	b := &APIBuilder{
		registerer:  registerer,
		pluginStore: pluginStore,
	}
	apiRegistrar.RegisterAPI(b)
	return b
}

func (b *APIBuilder) GetGroupVersion() schema.GroupVersion {
	return plugin.SchemeGroupVersion
}

func (b *APIBuilder) InstallSchema(scheme *runtime.Scheme) error {
	gv := plugin.SchemeGroupVersion
	//err := plugin.AddToScheme(scheme)
	//if err != nil {
	//	return err
	//}
	addKnownTypes(scheme, gv)

	// Link this version to the internal representation.
	// This is used for server-side-apply (PATCH), and avoids the error:
	//   "no kind is registered for the type"
	addKnownTypes(scheme, schema.GroupVersion{
		Group:   plugin.GROUP,
		Version: runtime.APIVersionInternal,
	})
	metav1.AddToGroupVersion(scheme, gv)
	return scheme.SetVersionPriority(gv)
}

func (b *APIBuilder) UpdateAPIGroupInfo(apiGroupInfo *genericapiserver.APIGroupInfo, opts builder.APIGroupOptions) error {
	resourceInfo := plugin.PluginResourceInfo
	s := map[string]rest.Storage{}

	storageReg, err := newPluginStorageWrapper(b.pluginStore, resourceInfo)
	if err != nil {
		return err
	}

	s[resourceInfo.StoragePath()] = storageReg

	apiGroupInfo.VersionedResourcesStorageMap[plugin.VERSION] = s
	return nil
}

func (b *APIBuilder) GetOpenAPIDefinitions() common.GetOpenAPIDefinitions {
	return plugin.GetOpenAPIDefinitions
}

func (b *APIBuilder) GetAuthorizer() authorizer.Authorizer {
	return authorizer.AuthorizerFunc(
		func(ctx context.Context, attr authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
			if !attr.IsResourceRequest() {
				return authorizer.DecisionNoOpinion, "", nil
			}

			u, err := identity.GetRequester(ctx)
			if err != nil {
				return authorizer.DecisionDeny, "valid user is required", err
			}

			if u.GetIsGrafanaAdmin() {
				return authorizer.DecisionAllow, "", nil
			}

			switch attr.GetVerb() {
			case "create":
				// Create requests are validated later since we don't have access to the resource name
				return authorizer.DecisionNoOpinion, "", nil
			case "get", "delete", "patch", "update", "list":
				return authorizer.DecisionAllow, "", nil
			default:
				// Forbid the rest
				return authorizer.DecisionDeny, "forbidden", nil
			}
		})
}

func addKnownTypes(scheme *runtime.Scheme, gv schema.GroupVersion) {
	scheme.AddKnownTypes(gv,
		&plugin.PluginResource{},
		&plugin.PluginResourceList{},
		&metav1.Status{},
	)
}
