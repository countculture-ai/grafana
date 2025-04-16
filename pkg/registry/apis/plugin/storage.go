package plugin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/grafana/grafana/pkg/apimachinery/utils"
	plugin "github.com/grafana/grafana/pkg/apis/plugin/v0alpha1"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
	gapiutil "github.com/grafana/grafana/pkg/services/apiserver/utils"
	"github.com/grafana/grafana/pkg/services/pluginsintegration/pluginstore"
)

var (
	_ rest.Scoper               = (*pluginStorageWrapper)(nil)
	_ rest.SingularNameProvider = (*pluginStorageWrapper)(nil)
	_ rest.Getter               = (*pluginStorageWrapper)(nil)
	_ rest.Lister               = (*pluginStorageWrapper)(nil)
	_ rest.Storage              = (*pluginStorageWrapper)(nil)
)

type pluginStorageWrapper struct {
	pluginStore pluginstore.Store

	resourceInfo   utils.ResourceInfo
	tableConverter rest.TableConvertor
}

func newPluginStorageWrapper(pluginStore pluginstore.Store, resourceInfo utils.ResourceInfo) (*pluginStorageWrapper, error) {
	return &pluginStorageWrapper{
		pluginStore:    pluginStore,
		resourceInfo:   resourceInfo,
		tableConverter: resourceInfo.TableConverter(),
	}, nil
}

func (s *pluginStorageWrapper) New() runtime.Object {
	return s.resourceInfo.NewFunc()
}

func (s *pluginStorageWrapper) Destroy() {}

func (s *pluginStorageWrapper) NamespaceScoped() bool {
	return true
}

func (s *pluginStorageWrapper) GetSingularName() string {
	return s.resourceInfo.GetSingularName()
}

func (s *pluginStorageWrapper) ShortNames() []string {
	return s.resourceInfo.GetShortNames()
}

func (s *pluginStorageWrapper) NewList() runtime.Object {
	return s.resourceInfo.NewListFunc()
}

func (s *pluginStorageWrapper) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return s.tableConverter.ConvertToTable(ctx, object, tableOptions)
}

func (s *pluginStorageWrapper) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	ns, err := request.NamespaceInfoFrom(ctx, true)
	if err != nil {
		return nil, err
	}

	p, exists := s.pluginStore.Plugin(ctx, name)
	if exists {
		return pluginToResource(p, ns.Value), nil
	}

	return nil, errors.New("plugin not found")
}

func (s *pluginStorageWrapper) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	ns, err := request.NamespaceInfoFrom(ctx, true)
	if err != nil {
		return nil, err
	}

	ps := s.pluginStore.Plugins(ctx)

	res := &plugin.PluginResourceList{
		Items: make([]plugin.PluginResource, len(ps)),
	}

	for i, p := range ps {
		res.Items[i] = *pluginToResource(p, ns.Value)
	}

	return res, nil
}

func pluginToResource(p pluginstore.Plugin, ns string) *plugin.PluginResource {
	ts := time.Now()

	pr := &plugin.PluginResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       plugin.KIND,
			APIVersion: plugin.VERSION,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              p.ID,
			Namespace:         ns,
			ResourceVersion:   fmt.Sprintf("%d", ts.UnixMilli()),
			CreationTimestamp: metav1.NewTime(ts),
		},
		Spec: plugin.PluginSpec{
			ID:      p.ID,
			Version: p.Info.Version,
		},
	}

	pr.UID = gapiutil.CalculateClusterWideUID(pr) // indicates if the value changed on the server
	meta, err := utils.MetaAccessor(pr)
	if err != nil {
		meta.SetUpdatedTimestamp(&ts)
	}
	return pr
}
