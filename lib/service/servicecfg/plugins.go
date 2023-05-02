// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package servicecfg

type PluginsConfig struct {
	Enabled bool
	// OpsgenieAPI is an Opsgenie API key.
	OpsgenieAPIKey string
	// Plugins is a map of labels used to match plugin resources.
	Plugins map[string]string
}

// IsEmpty validates if the Plugins Service config has no matchers.
func (d PluginsConfig) IsEmpty() bool {
	return len(d.Plugins) == 0
}
