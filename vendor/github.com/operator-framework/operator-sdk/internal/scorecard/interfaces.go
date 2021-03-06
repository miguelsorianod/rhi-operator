// Copyright 2019 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scorecard

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/operator-framework/operator-sdk/internal/scaffold"
	scplugins "github.com/operator-framework/operator-sdk/internal/scorecard/plugins"
	scapiv1alpha1 "github.com/operator-framework/operator-sdk/pkg/apis/scorecard/v1alpha1"
	"github.com/operator-framework/operator-sdk/version"
	v1 "k8s.io/api/core/v1"
)

type Plugin interface {
	List() scapiv1alpha1.ScorecardOutput
	Run() scapiv1alpha1.ScorecardOutput
}

type basicOrOLMPlugin struct {
	pluginType scplugins.PluginType
	config     scplugins.BasicAndOLMPluginConfig
}

func (p basicOrOLMPlugin) List() scapiv1alpha1.ScorecardOutput {
	res, err := scplugins.ListInternalPlugin(p.pluginType, p.config)
	if err != nil {
		Log.Errorf("%v", err)
		return scapiv1alpha1.ScorecardOutput{}
	}
	return res
}

func (p basicOrOLMPlugin) Run() scapiv1alpha1.ScorecardOutput {
	pluginLogs := &bytes.Buffer{}
	res, err := scplugins.RunInternalPlugin(p.pluginType, p.config, pluginLogs)
	if err != nil {
		var name string
		if p.pluginType == scplugins.BasicOperator {
			name = fmt.Sprintf("Basic Tests")
		} else if p.pluginType == scplugins.OLMIntegration {
			name = fmt.Sprintf("OLM Integration")
		}
		logs := fmt.Sprintf("%s:\nLogs: %s", err, pluginLogs.String())
		// output error to main logger as well for human-readable output
		Log.Errorf("Plugin `%s` failed with error (%v)", name, err)
		return failedPlugin(name, logs)
	}
	stderrString := pluginLogs.String()
	if len(stderrString) != 0 {
		Log.Warn(stderrString)
	}
	return res
}

// setConfigDefaults sets certain config fields to default values if they are not set
func setConfigDefaults(config *scplugins.BasicAndOLMPluginConfig, kubeconfig string) {
	if config.InitTimeout == 0 {
		config.InitTimeout = 60
	}
	if config.ProxyImage == "" {
		config.ProxyImage = fmt.Sprintf("quay.io/operator-framework/scorecard-proxy:%s", strings.TrimSuffix(version.Version, "+git"))
	}
	if config.ProxyPullPolicy == "" {
		config.ProxyPullPolicy = v1.PullAlways
	}
	if config.CRDsDir == "" {
		config.CRDsDir = scaffold.CRDsDir
	}
	if config.Kubeconfig == "" {
		config.Kubeconfig = kubeconfig
	}
}

func failedPlugin(name, log string) scapiv1alpha1.ScorecardOutput {
	return scapiv1alpha1.ScorecardOutput{
		Results: []scapiv1alpha1.ScorecardSuiteResult{{
			Name:  name,
			Error: 1,
			Log:   log,
		}},
	}
}
