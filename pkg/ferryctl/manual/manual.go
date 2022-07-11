package manual

import (
	"fmt"
	"strings"

	"github.com/ferry-proxy/ferry/pkg/consts"
)

type BuildManualPortConfig struct {
	ImportServiceName string

	BindPort   string
	ExportPort string
	ExportHost string

	ExportHubName string

	ExportTunnelHost     string
	ExportTunnelPort     string
	ExportTunnelIdentity string

	ImportTunnelHost     string
	ImportTunnelPort     string
	ImportTunnelIdentity string
}

func BuildManualPort(conf BuildManualPortConfig) (exportPortResource, importPortResource, importAddress string, err error) {
	namespace := consts.FerryTunnelNamespace
	exportName := fmt.Sprintf("%s-%s", conf.ImportServiceName, "export")
	importName := fmt.Sprintf("%s-%s", conf.ImportServiceName, "import")
	exportPortResource, err = buildExport(buildExportConfig{
		ExportName:      exportName,
		ExportNamespace: namespace,
		BindPort:        conf.BindPort,
		ExportPort:      conf.ExportPort,
		ExportHost:      conf.ExportHost,

		ImportTunnelHost:     conf.ImportTunnelHost,
		ImportTunnelPort:     conf.ImportTunnelPort,
		ImportTunnelIdentity: conf.ImportTunnelIdentity,
	})
	if err != nil {
		return "", "", "", err
	}
	importPortResource, err = buildImport(buildImportConfig{
		ImportServiceName: conf.ImportServiceName,
		ImportName:        importName,
		ImportNamespace:   namespace,
		BindPort:          conf.BindPort,
		ExportPort:        conf.ExportPort,
		ExportHost:        conf.ExportHost,

		ExportHubName: conf.ExportHubName,

		ExportTunnelHost:     conf.ExportTunnelHost,
		ExportTunnelPort:     conf.ExportTunnelPort,
		ExportTunnelIdentity: conf.ExportTunnelIdentity,
	})
	if err != nil {
		return "", "", "", err
	}

	exportPortResource = strings.TrimSpace(exportPortResource)
	importPortResource = strings.TrimSpace(importPortResource)

	importAddress = fmt.Sprintf("%s.%s.svc:%s", conf.ImportServiceName, namespace, conf.ExportPort)

	return exportPortResource, importPortResource, importAddress, nil
}
