// SPDX-License-Identifier: GPL-3.0-or-later

package nvme_raid

import (
	"fmt"
)

func (c *Collector) collect() (map[string]int64, error) {
	nvme_raidResp, err := c.queryRaidInfo()
	if err != nil {
		return nil, err
	}

	mx := make(map[string]int64)

	if err := c.collectRaidInfo(mx, nvme_raidResp); err != nil {
		return nil, fmt.Errorf("error collecting raid info: %s", err)
	}
	return mx, nil
}
