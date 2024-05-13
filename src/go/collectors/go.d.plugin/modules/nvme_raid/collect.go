// SPDX-License-Identifier: GPL-3.0-or-later

package nvme_raid

import (
	"fmt"
)

func (s *Nvme_Raid) collect() (map[string]int64, error) {
	nvme_raidResp, err := s.queryRaidInfo()
	if err != nil {
		return nil, err
	}

	raids := make(map[string]int64)

	if err := s.collectRaidInfo(raids, nvme_raidResp); err != nil {
		return nil, fmt.Errorf("error collecting raid info: %s", err)
	}
	return raids, nil
}
