package featuremgmt

import (
	"context"
	"fmt"
	"net/url"

	"github.com/grafana/grafana/pkg/setting"
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

const (
	StaticProviderType = "static"
	GOFFProviderType   = "goff"

	configSectionName  = "feature_toggles.openfeature"
	contextSectionName = "feature_toggles.openfeature.context"
)

type OpenFeatureService struct {
	provider            openfeature.FeatureProvider
	staticProviderFlags map[string]memprovider.InMemoryFlag

	Client openfeature.IClient

	ProviderType string
	URL          *url.URL
}

func ProvideOpenFeatureService(cfg *setting.Cfg) (*OpenFeatureService, error) {
	conf := cfg.Raw.Section(configSectionName)
	provType := conf.Key("provider").MustString(StaticProviderType)
	ofURL := conf.Key("url").MustString("")
	key := conf.Key("targetingKey").MustString(cfg.AppURL)

	u, err := url.Parse(ofURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %s: %w", ofURL, err)
	}

	// TODO: move somehow to static provider implementation
	flags := make(map[string]memprovider.InMemoryFlag)

	var provider openfeature.FeatureProvider
	if provType == GOFFProviderType {
		provider, err = newGOFFProvider(ofURL)
	} else {
		provider, flags, err = newStaticProvider(cfg)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s feature provider: %w", provType, err)
	}

	if err := openfeature.SetProviderAndWait(provider); err != nil {
		return nil, fmt.Errorf("failed to set global %s feature provider: %w", provType, err)
	}

	attrs := ctxAttrs(cfg)
	openfeature.SetEvaluationContext(openfeature.NewEvaluationContext(key, attrs))

	client := openfeature.NewClient("grafana-openfeature-client")
	return &OpenFeatureService{
		provider:            provider,
		staticProviderFlags: flags,
		Client:              client,
		URL:                 u,
		ProviderType:        provType,
	}, nil
}

// ctxAttrs uses config.ini [feature_toggles.openfeature.context] section to build the eval context attributes
func ctxAttrs(cfg *setting.Cfg) map[string]any {
	ctxConf := cfg.Raw.Section(contextSectionName)

	attrs := map[string]any{}
	for _, key := range ctxConf.KeyStrings() {
		attrs[key] = ctxConf.Key(key).String()
	}

	// Some default attributes
	if _, ok := attrs["grafana_version"]; !ok {
		attrs["grafana_version"] = setting.BuildVersion
	}

	return attrs
}

func (s *OpenFeatureService) EvalFlagWithStaticProvider(ctx context.Context, flagKey string) (openfeature.BooleanEvaluationDetails, error) {
	if s.ProviderType == GOFFProviderType {
		return openfeature.BooleanEvaluationDetails{}, fmt.Errorf("request must be sent to open feature service for %s provider", GOFFProviderType)
	}

	result, err := s.Client.BooleanValueDetails(ctx, flagKey, false, openfeature.TransactionContext(ctx))
	if err != nil {
		return openfeature.BooleanEvaluationDetails{}, fmt.Errorf("failed to evaluate flag %s: %w", flagKey, err)
	}

	return result, nil
}

func (s *OpenFeatureService) EvalAllFlagsWithStaticProvider(ctx context.Context) (AllFlagsGOFFResp, error) {
	if s.ProviderType == GOFFProviderType {
		return AllFlagsGOFFResp{}, fmt.Errorf("request must be sent to open feature service for %s provider", GOFFProviderType)
	}

	// TODO: implement this

	return AllFlagsGOFFResp{}, nil
}

type AllFlagsGOFFResp struct {
	Flags map[string]*FlagGOFF `json:"flags"`
}

type FlagGOFF struct {
	VariationType string `json:"variationType"`
	Timestamp     int    `json:"timestamp"`
	TrackEvents   bool   `json:"trackEvents"`
	Value         bool   `json:"value"`
}
