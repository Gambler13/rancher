package samlconfig

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rancher/rancher/pkg/auth/providers/saml"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type authProvider struct {
	authConfigs v3.AuthConfigInterface
}

func Register(apiContext *config.ScaledContext) {
	a := newAuthProvider(apiContext)
	apiContext.Management.AuthConfigs("").AddHandler("authConfigController", a.sync)
}

func newAuthProvider(apiContext *config.ScaledContext) *authProvider {
	a := &authProvider{
		authConfigs: apiContext.Management.AuthConfigs(""),
	}
	return a
}

func (a *authProvider) sync(key string, config *v3.AuthConfig) error {
	samlConfig := &v3.SamlConfig{}
	if key == "" || config == nil {
		return nil
	}

	if config.Name != saml.PingName {
		return nil
	}

	if config.Annotations["configured"] != "true" {
		return nil
	}

	authConfigObj, err := a.authConfigs.ObjectClient().UnstructuredClient().Get(config.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to retrieve SamlConfig, error: %v", err)
	}

	u, ok := authConfigObj.(runtime.Unstructured)
	if !ok {
		return fmt.Errorf("failed to retrieve SamlConfig, cannot read k8s Unstructured data")
	}
	storedSamlConfigMap := u.UnstructuredContent()
	decode(storedSamlConfigMap, samlConfig)

	metadataMap, ok := storedSamlConfigMap["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to retrieve SamlConfig metadata, cannot read k8s Unstructured data")
	}

	typemeta := &metav1.ObjectMeta{}
	mapstructure.Decode(metadataMap, typemeta)
	samlConfig.ObjectMeta = *typemeta

	return saml.InitializeSamlClient(samlConfig, config.Name)
}

func decode(m interface{}, rawVal interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   rawVal,
		TagName:  "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(m)
}