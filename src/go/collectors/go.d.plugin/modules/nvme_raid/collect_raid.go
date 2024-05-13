package nvme_raid

import (
	"encoding/json"
	"errors"
)

type (
	nvme_RaidInfoResponse struct {
		Raids []struct {
			Raid_Data raid_data `json:"raid_data"`
		} `json:"raids"`
	}
	raid_data struct {
		Active     bool `json:"active"`
		Block_Size int  `json:"block_size"`
		Config     bool `json:"config"`
		Devices    []struct {
			ID     int      `json:"id"`
			Device string   `json:"device"`
			Status []string `json:"status"`
		} `json:"devices"`
		Devices_Health      []string `json:"devices_health"`
		Devices_Wear        []string `json:"devices_wear"`
		Group_Size          int      `json:"group_size"`
		Init_Depth          int      `json:"init_depth"`
		Init_Prio           int      `json:"init_prio"`
		Level               string   `json:"level"`
		Memory_Limit_Mb     int      `json:"memory_limit_mb"`
		Memory_Usage_Mb     string   `json:"memory_usage_mb"`
		Merge_Max_Usecs     int      `json:"merge_max_usecs"`
		Merge_Read_Enabled  int      `json:"merge_read_enabled"`
		Merge_Wait_Usecs    int      `json:"merge_wait_usecs"`
		Merge_Write_Enabled int      `json:"merge_write_enabled"`
		Name                string   `json:"name"`
		Recon_Depth         int      `json:"recon_depth"`
		Recon_Prio          int      `json:"recon_prio"`
		Request_Limit       int      `json:"request_limit"`
		Restripe_Prio       int      `json:"restripe_prio"`
		Resync_Enabled      int      `json:"resync_enabled"`
		Sched_Enabled       int      `json:"sched_enabled"`
		Serials             []string `json:"serials"`
		Size                string   `json:"size"`
		Sparepool           string   `json:"sparepool"`
		State               []string `json:"state"`
		Strip_Size          int      `json:"strip_size"`
		UUID                string   `json:"uuid"`
	}
)

func (s *Nvme_Raid) collectRaidInfo(raid map[string]int64, resp *nvme_RaidInfoResponse) error {
	for _, raid := range resp.Raids {
		// Assuming there's only one RAID per response
		raidData := raid.Raid_Data
		if !s.raids[raidData.Name] {
			s.raids[raidData.Name] = true
			s.addRaidDataCharts(raidData)
		}
		// Example logic:
		// Generate device labels
	}

	return nil
}

// func (s *Nvme_Raid) collectDeviceInfo(raidData raid_data) {
// 	for i, device := range raidData.Devices {
// 		s.addDeviceCharts(i, device)
// 	}
// }

// func (s *Nvme_Raid) collectRaidInfo(raids map[string]int64, resp *Nvme_RaidInfoResponse) error {
// 	for _, raid := range resp.Raids {
// 		raidData := raid.Raid_Data
// 		fmt.Printf("Raid Name: %s\n", raidData.Name)
// 		fmt.Printf("Active: %t\n", raidData.Active)
// 		fmt.Printf("Block Size: %d\n", raidData.Block_Size)
// 		fmt.Printf("Config: %t\n", raidData.Config)
// 		fmt.Println("Devices:")
// 		for _, device := range raidData.Devices {
// 			fmt.Printf("  ID: %d, Device: %s, Status: %v\n", device.ID, device.Device, device.Status)
// 		}
// 		fmt.Printf("Devices Health: %v\n", raidData.Devices_Health)
// 		fmt.Printf("Devices Wear: %v\n", raidData.Devices_Wear)
// 		fmt.Printf("Group Size: %d\n", raidData.Group_Size)
// 		fmt.Printf("Init Depth: %d\n", raidData.Init_Depth)
// 		fmt.Printf("Init Prio: %d\n", raidData.Init_Prio)
// 		fmt.Printf("Level: %s\n", raidData.Level)
// 		fmt.Printf("Memory Limit Mb: %d\n", raidData.Memory_Limit_Mb)
// 		fmt.Printf("Memory Usage Mb: %s\n", raidData.Memory_Usage_Mb)
// 		fmt.Printf("Merge Max Usecs: %d\n", raidData.Merge_Max_Usecs)
// 		fmt.Printf("Merge Read Enabled: %d\n", raidData.Merge_Read_Enabled)
// 		fmt.Printf("Merge Wait Usecs: %d\n", raidData.Merge_Wait_Usecs)
// 		fmt.Printf("Merge Write Enabled: %d\n", raidData.Merge_Write_Enabled)
// 		fmt.Printf("Name: %s\n", raidData.Name)
// 		fmt.Printf("Recon Depth: %d\n", raidData.Recon_Depth)
// 		fmt.Printf("Recon Prio: %d\n", raidData.Recon_Prio)
// 		fmt.Printf("Request Limit: %d\n", raidData.Request_Limit)
// 		fmt.Printf("Restripe Prio: %d\n", raidData.Restripe_Prio)
// 		fmt.Printf("Resync Enabled: %d\n", raidData.Resync_Enabled)
// 		fmt.Printf("Sched Enabled: %d\n", raidData.Sched_Enabled)
// 		fmt.Printf("Serials: %v\n", raidData.Serials)
// 		fmt.Printf("Size: %s\n", raidData.Size)
// 		fmt.Printf("Sparepool: %s\n", raidData.Sparepool)
// 		fmt.Printf("State: %v\n", raidData.State)
// 		fmt.Printf("Strip Size: %d\n", raidData.Strip_Size)
// 		fmt.Printf("UUID: %s\n\n", raidData.UUID)
// 	}

// 	// Add any other processing or metrics collection for RAID configurations here

// 	return nil
// }

func (s *Nvme_Raid) queryRaidInfo() (*nvme_RaidInfoResponse, error) {
	bs, err := s.exec.nvme_raidInfo()
	if err != nil {
		return nil, err
	}

	if len(bs) == 0 {
		return nil, errors.New("empty response")
	}

	var resp nvme_RaidInfoResponse
	if err := json.Unmarshal(bs, &resp); err != nil {
		return nil, err
	}

	// Check if RAID configurations are present
	if len(resp.Raids) == 0 {
		return nil, errors.New("no RAID configurations found")
	}
	return &resp, nil
}
