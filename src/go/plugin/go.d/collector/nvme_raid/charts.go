// SPDX-License-Identifier: GPL-3.0-or-later

//go:build linux || freebsd || openbsd || netbsd || dragonfly

package nvme_raid

import (
	"fmt"

	"github.com/netdata/netdata/go/plugins/plugin/go.d/agent/module"
)

const (
	prioRaidData = module.Priority + iota
	prioDevice
)

var raidDataChartsTmpl = module.Charts{
	raidStateChartTmpl.Copy(),
	// Add more chart templates as needed...
}

var (
	raidStateChartTmpl = module.Chart{
		ID:       "raid_%s_state",
		Title:    "RAID State",
		Units:    "status",
		Fam:      "raid state",
		Ctx:      "nvme_raid.raid_state",
		Type:     module.Line,
		Priority: prioRaidData,
		Dims: module.Dims{
			// Define dimensions for different RAID states
			{ID: "raid_%s_state_online", Name: "online"},
			{ID: "raid_%s_state_initialized", Name: "initialized"},
			{ID: "raid_%s_state_initing", Name: "initing"},
			{ID: "raid_%s_state_degraded", Name: "degraded"},
			{ID: "raid_%s_state_reconstructing", Name: "reconstructing"},
			{ID: "raid_%s_state_offline", Name: "offline"},
			{ID: "raid_%s_state_need_recon", Name: "need_recon"},
			{ID: "raid_%s_state_need_init", Name: "need_init"},
			{ID: "raid_%s_state_read_only", Name: "read Only"},
			{ID: "raid_%s_state_unrecovered", Name: "unrecovered"},
			{ID: "raid_%s_state_none", Name: "none"},
			{ID: "raid_%s_state_restriping", Name: "restriping"},
			{ID: "raid_%s_state_need_resize", Name: "need_resize"},
			{ID: "raid_%s_state_need_restripe", Name: "need_restripe"},
		},
	}
	// Add more chart templates as needed...
)

// var raidDataChartsTmpl = module.Charts{
// 	raidDataChartTmpl.Copy(),
// }

// var (
// 	raidDataChartTmpl = module.Chart{
// 		ID:       "raid_data_%s",
// 		Title:    "Raid Data",
// 		Units:    "status",
// 		Fam:      "raid data",
// 		Ctx:      "nvme_raid.raid_data",
// 		Type:     module.Line,
// 		Priority: prioRaidData,
// 		Dims: module.Dims{
// 			{ID: "raid_data_%s_active", Name: "active"},
// 			{ID: "raid_data_%s_config", Name: "config"},
// 			// Add more dimensions as needed
// 		},
// 	}
// 	deviceChartsTmpl = module.Charts{
// 		deviceChartTmpl.Copy(),
// 	}

// 	devicehartTmpl = module.Chart{
// 		ID:       "device_%s_%d_status",
// 		Title:    "Device Status",
// 		Units:    "status",
// 		Fam:      "device status",
// 		Ctx:      "nvme_raid.device_status",
// 		Type:     module.Line,
// 		Priority: prioDevice,
// 		Dims: module.Dims{
// 			{ID: "device_%s_%d_status_online", Name: "online"},
// 			{ID: "device_%s_%d_status_offline", Name: "offline"},
// 			// Add more dimensions as needed
// 		},
// 	}
// )

func (c *Collector) addRaidDataCharts(raids raidData) {
	charts := raidDataChartsTmpl.Copy()

	// raidName := raids.Name

	for _, chart := range *charts {
		chart.ID = fmt.Sprintf(chart.ID, raids.Name)
		chart.Labels = []module.Label{
			{Key: "raid_name", Value: raids.Name},
			// You can add more labels here if needed
		}
		for _, dim := range chart.Dims {
			dim.ID = fmt.Sprintf(dim.ID, raids.Name)
		}
	}

	if err := c.Charts().Add(*charts...); err != nil {
		c.Warning(err)
	}
}
