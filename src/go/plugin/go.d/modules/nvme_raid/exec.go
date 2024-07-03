// SPDX-License-Identifier: GPL-3.0-or-later

package nvme_raid

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/netdata/netdata/go/plugins/logger"
)

func newNvme_RaidExec(ndsudoPath string, timeout time.Duration, log *logger.Logger) *nvme_RaidExec {
	return &nvme_RaidExec{
		Logger:     log,
		ndsudoPath: ndsudoPath,
		timeout:    timeout,
	}
}

type nvme_RaidExec struct {
	*logger.Logger

	ndsudoPath string
	timeout    time.Duration
}

// raidInfo implements nvme_Raid.

func (e *nvme_RaidExec) nvme_raidInfo() ([]byte, error) {
	return e.execute("nvme_raid-show")
}

func (e *nvme_RaidExec) execute(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.ndsudoPath, args...)
	e.Debugf("executing '%s'", cmd)

	bs, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error on '%s': %v", cmd, err)
	}

	return bs, nil
}
