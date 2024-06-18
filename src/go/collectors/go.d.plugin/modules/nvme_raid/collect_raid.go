package nvme_raid

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type (
	// Define structs to model JSON responses
	nvme_RaidInfoResponse struct {
		Raids map[string]raid_data `json:"-"`
	}
	raid_data struct {
		Active              bool     `json:"active"`
		Block_Size          int      `json:"block_size"`
		Config              bool     `json:"config"`
		Devices             []device `json:"devices"`
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
	device struct {
		ID     int      `json:"-"`
		Device string   `json:"-"`
		Status []string `json:"-"`
	}
)

func (r *raid_data) UnmarshalJSON(data []byte) error {
	type Alias raid_data // Define an alias to prevent infinite recursion
	aux := &struct {
		Devices [][]interface{} `json:"devices"` // Use interface{} to handle mixed types
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
			return fmt.Errorf("unexpected device format")
		}

		id, ok := dev[0].(float64) // JSON numbers are float64 by default
		if !ok {
			return fmt.Errorf("unexpected type for device ID")
		}
		devicePath, ok := dev[1].(string)
		if !ok {
			return fmt.Errorf("unexpected type for device path")
		}
		status, ok := dev[2].([]interface{})
		if !ok {
			return fmt.Errorf("unexpected type for device status")
		}
		statusStr := make([]string, len(status))
		for j, s := range status {
			statusStr[j], ok = s.(string)
			if !ok {
				return fmt.Errorf("unexpected type in device status")
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

func (s *Nvme_Raid) collectRaidInfo(raids map[string]int64, resp *nvme_RaidInfoResponse) error {
	for _, raid := range resp.Raids {
		raidData := raid
		raidName := raidData.Name

		// Check if the RAID has already been processed
		if !s.raids[raidName] {
			// Mark RAID as processed
			s.raids[raidName] = true
			// Add RAID data charts
			s.addRaidDataCharts(raidData)
		}

		// Create prefix for metrics related to this RAID
		px := fmt.Sprintf("raid_%s_", raidName)
		// Define RAID states
		raidStates := []string{
			"online", "initialized", "initing", "degraded", "reconstructing",
			"offline", "need_recon", "need_init", "read Only", "unrecovered",
			"none", "restriping", "need_resize", "need_restripe",
		}

		// Initialize metrics for common RAID states
		for _, state := range raidStates {
			// Set the initial value of each state metric to 0
			raids[px+"state_"+state] = 0
		}
		// Switch statement to handle different numbers of state types
		// Note: strings.ToLower is used to ensure consistency in the metric keys
		switch len(raidData.State) {
		case 1:
			// If there is only one RAID state, set its corresponding metric to 1
			state := strings.ToLower(raidData.State[0])
			raids[px+"state_"+state] = 1
		case 2:
			// If there are two RAID states, set the corresponding metrics to 1
			state1 := strings.ToLower(raidData.State[0])
			state2 := strings.ToLower(raidData.State[1])
			raids[px+"state_"+state1] = 1 // Set the first state metric to 1
			raids[px+"state_"+state2] = 1 // Set the second state metric to 1
		case 3:
			// If there are three RAID states, set the corresponding metrics to 1
			state1 := strings.ToLower(raidData.State[0])
			state2 := strings.ToLower(raidData.State[1])
			state3 := strings.ToLower(raidData.State[2])
			raids[px+"state_"+state1] = 1 // Set the first state metric to 1
			raids[px+"state_"+state2] = 1 // Set the second state metric to 1
			raids[px+"state_"+state3] = 1 // Set the third state metric to 1
		default:
			// Handle the case where the number of states is unexpected
			return errors.New("unexpected number of states")
		}
	}
	return nil
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

func (s *Nvme_Raid) queryRaidInfo() (*nvme_RaidInfoResponse, error) {
	// Call the exec method to retrieve RAID information
	bs, err := s.exec.nvme_raidInfo()
	if err != nil {
		return nil, fmt.Errorf("error retrieving RAID info: %v", err)
	}

	// Check if the response is empty
	if len(bs) == 0 {
		return nil, errors.New("empty response")
	}

	// Define a struct to hold the JSON response
	var resp nvme_RaidInfoResponse

	// Unmarshal the JSON byte slice into the resp struct
	if err := json.Unmarshal(bs, &resp); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// Check if there are RAID configurations present in the response
	if len(resp.Raids) == 0 {
		return nil, errors.New("no RAID configurations found")
	}

	// Return the parsed response and no error
	return &resp, nil
}
