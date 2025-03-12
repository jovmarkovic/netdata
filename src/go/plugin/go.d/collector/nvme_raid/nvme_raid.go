// SPDX-License-Identifier: GPL-3.0-or-later

package nvme_raid

import (
	_ "embed"
	"errors"
	"time"

	"github.com/netdata/netdata/go/plugins/plugin/go.d/agent/module"
	"github.com/netdata/netdata/go/plugins/plugin/go.d/pkg/web"
)

//go:embed "config_schema.json"
var configSchema string

func init() {
	module.Register("nvme_raid", module.Creator{
		JobConfigSchema: configSchema,
		Defaults: module.Defaults{
			UpdateEvery: 10,
		},
		Create: func() module.Module { return New() },
	})
}

func New() *Nvme_Raid {
	return &Nvme_Raid{
		Config: Config{
			Timeout: web.Duration(time.Second * 2),
		},
		charts: &module.Charts{},
		raids:  make(map[string]bool),
	}
}

type Config struct {
	UpdateEvery int          `yaml:"update_every" json:"update_every"`
	Timeout     web.Duration `yaml:"timeout" json:"timeout"`
}

type (
	Nvme_Raid struct {
		module.Base
		Config `yaml:",inline" json:""`

		charts *module.Charts

		exec nvme_Raid

		raids map[string]bool
	}
	nvme_Raid interface {
		nvme_raidInfo() ([]byte, error)
	}
)

func (s *Nvme_Raid) Configuration() any {
	return s.Config
}

func (s *Nvme_Raid) Init() error {
	nvme_raidExec, err := s.initNvme_RaidExec()
	if err != nil {
		s.Errorf("nvme_raid exec initialization: %v", err)
		return err
	}
	s.exec = nvme_raidExec

	return nil
}

func (s *Nvme_Raid) Check() error {
	raids, err := s.collect()
	if err != nil {
		s.Error(err)
		return err
	}

	if len(raids) == 0 {
		return errors.New("no metrics collected")
	}

	return nil
}

func (s *Nvme_Raid) Charts() *module.Charts {
	return s.charts
}

func (s *Nvme_Raid) Collect() map[string]int64 {
	raids, err := s.collect()
	if err != nil {
		s.Error(err)
	}

	if len(raids) == 0 {
		return nil
	}

	return raids
}

func (s *Nvme_Raid) Cleanup() {}
