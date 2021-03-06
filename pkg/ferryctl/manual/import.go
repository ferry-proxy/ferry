/*
Copyright 2022 FerryProxy Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manual

import (
	_ "embed"

	"github.com/ferryproxy/ferry/pkg/ferryctl/utils"
)

type buildImportConfig struct {
	ImportServiceName string
	ImportName        string
	ImportNamespace   string
	BindPort          string
	ExportPort        string
	ExportHost        string
	ExportHubName     string

	ExportTunnelHost     string
	ExportTunnelPort     string
	ExportTunnelIdentity string
}

func buildImport(conf buildImportConfig) (string, error) {
	return utils.RenderString(importYaml, conf), nil
}

//go:embed import.yaml
var importYaml string
