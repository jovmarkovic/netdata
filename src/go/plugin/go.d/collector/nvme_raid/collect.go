// SPDX-License-Identifier: GPL-3.0-or-later

//go:build linux || freebsd || openbsd || netbsd || dragonfly

package nvme_raid

import (
	"fmt"
)

func (c *Collector) collect() (map[string]int64, error) {
	nvmeRaidResp, err := c.queryRaidInfo()
	if err != nil {
		return nil, err
	}

	mx := make(map[string]int64)

	if err := c.collectRaidInfo(mx, nvmeRaidResp); err != nil {
		return nil, fmt.Errorf("error collecting raid info: %s", err)
	}
	return mx, nil
}
