package init

import (
	"fmt"

	"github.com/ferryproxy/ferry/pkg/ferryctl/control_plane"
	"github.com/ferryproxy/ferry/pkg/ferryctl/data_plane"
	"github.com/ferryproxy/ferry/pkg/ferryctl/kubectl"
	"github.com/ferryproxy/ferry/pkg/ferryctl/log"
	"github.com/ferryproxy/ferry/pkg/ferryctl/vars"
	"github.com/spf13/cobra"
)

func NewCommand(logger log.Logger) *cobra.Command {
	var (
		controlPlaneTunnelAddress = vars.AutoPlaceholders
		controlPlaneReachable     = true
	)

	cmd := &cobra.Command{
		Use:  "init",
		Args: cobra.NoArgs,
		Aliases: []string{
			"i",
		},
		Short: "Control plane init commands",
		Long:  `Control plane init commands is used to initialize the control plane`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("too many arguments")
			}

			kctl := kubectl.NewKubectl()
			var err error
			if controlPlaneTunnelAddress == vars.AutoPlaceholders {
				controlPlaneTunnelAddress, err = kctl.GetTunnelAddress(cmd.Context())
				if err != nil {
					return err
				}
			}

			err = data_plane.ClusterInit(cmd.Context(), data_plane.ClusterInitConfig{
				FerryTunnelImage: vars.FerryTunnelImage,
			})
			if err != nil {
				return err
			}

			err = control_plane.ClusterInit(cmd.Context(), control_plane.ClusterInitConfig{
				ControlPlaneName:          vars.ControlPlaneName,
				ControlPlaneReachable:     controlPlaneReachable,
				ControlPlaneTunnelAddress: controlPlaneTunnelAddress,
				FerryControllerImage:      vars.FerryControllerImage,
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&controlPlaneTunnelAddress, "control-plane-tunnel-address", controlPlaneTunnelAddress, "Tunnel address of the control plane connected to another cluster")
	flags.BoolVar(&controlPlaneReachable, "control-plane-reachable", controlPlaneReachable, "Whether the control plane is reachable")
	return cmd
}
