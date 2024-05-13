package nvme_raid

import (
	"fmt"

	"github.com/netdata/netdata/go/go.d.plugin/agent/module"
)

const (
	prioRaidData = module.Priority + iota
	prioDevice
)

var raidDataChartsTmpl = module.Charts{
	raidDataChartTmpl.Copy(),
}

var (
	raidDataChartTmpl = module.Chart{
		ID:       "raid_data_%s",
		Title:    "Raid Data",
		Units:    "status",
		Fam:      "raid data",
		Ctx:      "nvme_raid.raid_data",
		Type:     module.Line,
		Priority: prioRaidData,
		Dims: module.Dims{
			{ID: "raid_data_%s_active", Name: "active"},
			{ID: "raid_data_%s_config", Name: "config"},
			// Add more dimensions as needed
		},
	}
	deviceChartsTmpl = module.Charts{
		deviceChartTmpl.Copy(),
	}

	deviceChartTmpl = module.Chart{
		ID:       "device_%s_%d_status",
		Title:    "Device Status",
		Units:    "status",
		Fam:      "device status",
		Ctx:      "nvme_raid.device_status",
		Type:     module.Line,
		Priority: prioDevice,
		Dims: module.Dims{
			{ID: "device_%s_%d_status_online", Name: "online"},
			{ID: "device_%s_%d_status_offline", Name: "offline"},
			// Add more dimensions as needed
		},
	}
)

func (s *Nvme_Raid) addRaidDataCharts(raidData raid_data) {
	charts := raidDataChartsTmpl.Copy()

	for _, chart := range *charts {
		chart.ID = fmt.Sprintf(chart.ID, raidData.Name)
		chart.Labels = []module.Label{
			{Key: "raid_data_name", Value: raidData.Name},
			// Add more labels as needed
		}
		for _, dim := range chart.Dims {
			dim.ID = fmt.Sprintf(dim.ID, raidData.Name)
		}
	}

	if err := s.Charts().Add(*charts...); err != nil {
		s.Warning(err)
	}
}
