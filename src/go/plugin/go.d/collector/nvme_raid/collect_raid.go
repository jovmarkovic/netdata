// SPDX-License-Identifier: GPL-3.0-or-later

//go:build linux || freebsd || openbsd || netbsd || dragonfly

package nvme_raid

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type (
	// Define structs to model JSON responses
	nvmeRaidInfoResponse struct {
		Raids map[string]raidData `json:"-"` // Use map to dynamically hold RAID configurations
	}

	raidData struct {
		Active            bool     `json:"active"`
		BlockSize         int      `json:"block_size"`
		Config            bool     `json:"config"`
		Devices           []device `json:"devices"`
		DevicesHealth     []string `json:"devices_health"`
		DevicesWear       []string `json:"devices_wear,omitempty"` // Handle optional field
		GroupSize         int      `json:"group_size"`
		InitDepth         int      `json:"init_depth"`
		InitPrio          int      `json:"init_prio"`
		Level             string   `json:"level"`
		MemoryLimitMb     int      `json:"memory_limit_mb"`
		MemoryUsageMb     string   `json:"memory_usage_mb"`
		MergeMaxUsecs     int      `json:"merge_max_usecs"`
		MergeReadEnabled  int      `json:"merge_read_enabled"`
		MergeWaitUsecs    int      `json:"merge_wait_usecs"`
		MergeWriteEnabled int      `json:"merge_write_enabled"`
		Name              string   `json:"name"`
		ReconDepth        int      `json:"recon_depth"`
		ReconPrio         int      `json:"recon_prio"`
		RequestLimit      int      `json:"request_limit"`
		RestripePrio      int      `json:"restripe_prio"`
		ResyncEnabled     int      `json:"resync_enabled"`
		SchedEnabled      int      `json:"sched_enabled"`
		Serials           []string `json:"serials"`
		Size              string   `json:"size"`
		Sparepool         string   `json:"sparepool"`
		State             []string `json:"state"`
		StripSize         int      `json:"strip_size"`
		UUID              string   `json:"uuid"`
	}
	device struct {
		ID     int      `json:"-"`
		Device string   `json:"-"`
		Status []string `json:"-"`
	}
)

func (r *raidData) UnmarshalJSON(data []byte) error {
	type Alias raidData // Define an alias to prevent infinite recursion
	aux := &struct {
		Devices [][]any `json:"devices"` // Use any to handle mixed types
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert the parsed data into Device structs
	r.Devices = make([]device, len(aux.Devices))
	for i, dev := range aux.Devices {
		if len(dev) != 3 {
			return fmt.Errorf("invalid device format at index %d", i)
		}
		id, ok := dev[0].(float64) // JSON numbers are float64 by default
		if !ok {
			return fmt.Errorf("device ID at index %d should be a number", i)
		}
		devicePath, ok := dev[1].(string)
		if !ok {
			return fmt.Errorf("device path at index %d should be a string", i)
		}
		status, ok := dev[2].([]any)
		if !ok {
			return fmt.Errorf("device status at index %d should be a list", i)
		}
		statusStr := make([]string, len(status))
		for j, s := range status {
			if str, ok := s.(string); ok {
				statusStr[j] = str
			} else {
				return fmt.Errorf("device status entry at index [%d][%d] should be a string", i, j)
			}
		}

		r.Devices[i] = device{
			ID:     int(id),
			Device: devicePath,
			Status: statusStr,
		}
	}

	return nil
}

func (c *Collector) collectRaidInfo(raids map[string]int64, resp *nvmeRaidInfoResponse) error {
	for _, raid := range resp.Raids {
		// Mark RAID as processed if it hasn't been processed yet
		if !c.raids[raid.Name] {
			c.raids[raid.Name] = true
			c.addRaidDataCharts(raid)
		}

		// Create prefix for RAID-related metrics
		px := fmt.Sprintf("raid_%s_", raid.Name)

		// Initialize all possible RAID state metrics to 0
		for _, st := range []string{
			"online", "initialized", "initing", "degraded", "reconstructing",
			"offline", "need_recon", "need_init", "read_only", "unrecovered",
			"none", "restriping", "need_resize", "need_restripe",
		} {
			raids[px+"state_"+st] = 0
		}

		// Set metrics for RAID's current states
		for _, st := range raid.State {
			raids[px+"state_"+strings.ToLower(st)] = 1
		}
	}

	return nil
}

func (c *Collector) queryRaidInfo() (*nvmeRaidInfoResponse, error) {
	// Call the exec method to retrieve RAID information
	bs, err := c.exec.nvmeRaidInfo()
	if err != nil {
		return nil, fmt.Errorf("error retrieving RAID info: %v", err)
	}

	// Check if the response is empty
	if len(bs) == 0 {
		return nil, errors.New("empty response")
	}

	// Create a temporary map to hold the data
	tempMap := make(map[string]json.RawMessage)
	if err := json.Unmarshal(bs, &tempMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// Initialize the response
	resp := nvmeRaidInfoResponse{
		Raids: make(map[string]raidData),
	}

	// Unmarshal each RAID individually
	for key, value := range tempMap {
		var rd raidData
		if err := json.Unmarshal(value, &rd); err != nil {
			return nil, fmt.Errorf("error unmarshalling RAID data for %s: %v", key, err)
		}
		resp.Raids[key] = rd
	}

	// Check if there are RAID configurations present in the response
	if len(resp.Raids) == 0 {
		return nil, errors.New("no RAID configurations found")
	}

	// Return the parsed response and no error
	return &resp, nil
}

// 		// Example logic:
// 		// Generate device labels
// 		for i, device := range raidData.Devices {
// 			// Assuming you have a function to add device charts
// 			s.addDeviceCharts(i, device)

// 			// Construct the metric name using the prefix 'px' and the device index 'i'
// 			// and add it to the 'raids' map
// 			raids[px+"device_"+strconv.Itoa(i)] = 1 // You can set a value based on your requirements
// 		}
// 	}

// 	return nil
// }

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
