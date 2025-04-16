package plugin

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/rest"

	grafanaregistry "github.com/grafana/grafana/pkg/apiserver/registry/generic"
)

type genericStrategy interface {
	rest.RESTCreateStrategy
	rest.RESTUpdateStrategy
}

var _ rest.RESTCreateStrategy = (*pluginStorageStrategy)(nil)

type pluginStorageStrategy struct {
	genericStrategy

	registerer prometheus.Registerer
}

func newStrategy(typer runtime.ObjectTyper, gv schema.GroupVersion, registerer prometheus.Registerer) *pluginStorageStrategy {
	return &pluginStorageStrategy{grafanaregistry.NewStrategy(typer, gv), registerer}
}

func (s *pluginStorageStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (s *pluginStorageStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}
