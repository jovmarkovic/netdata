// SPDX-License-Identifier: GPL-3.0-or-later

package nvme_raid

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/netdata/netdata/go/plugins/pkg/executable"
)

func (s *Nvme_Raid) initNvme_RaidExec() (nvme_Raid, error) {
	ndsudoPath := filepath.Join(executable.Directory, "ndsudo")

	if _, err := os.Stat(ndsudoPath); err != nil {
		return nil, fmt.Errorf("ndsudo executable not found: %v", err)
	}

	nvme_raidExec := newNvme_RaidExec(ndsudoPath, s.Timeout.Duration(), s.Logger)

	return nvme_raidExec, nil
}
