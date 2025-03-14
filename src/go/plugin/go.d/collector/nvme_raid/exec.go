// SPDX-License-Identifier: GPL-3.0-or-later

//go:build linux || freebsd || openbsd || netbsd || dragonfly

package nvme_raid

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/netdata/netdata/go/plugins/logger"
)

type nvmeRaid interface {
	nvmeRaidInfo() ([]byte, error)
}

func newNvmeRaidExec(ndsudoPath string, timeout time.Duration, log *logger.Logger) *nvmeRaidExec {
	return &nvmeRaidExec{
		Logger:     log,
		ndsudoPath: ndsudoPath,
		timeout:    timeout,
	}
}

type nvmeRaidExec struct {
	*logger.Logger

	ndsudoPath string
	timeout    time.Duration
}

func (e *nvmeRaidExec) nvmeRaidInfo() ([]byte, error) {
	return e.execute("nvme_raid-show")
}

func (e *nvmeRaidExec) execute(args ...string) ([]byte, error) {
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
