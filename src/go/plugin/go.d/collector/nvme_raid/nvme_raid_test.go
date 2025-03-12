// SPDX-License-Identifier: GPL-3.0-or-later

package nvme_raid

import (
	"errors"
	"os"
	"testing"

	"github.com/netdata/netdata/go/plugins/plugin/go.d/agent/module"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dataConfigJSON, _ = os.ReadFile("testdata/config.json")
	dataConfigYAML, _ = os.ReadFile("testdata/config.yaml")

	nvme_raidData, _ = os.ReadFile("testdata/nvme_raid_data.json")
)

func Test_testDataIsValid(t *testing.T) {
	for name, data := range map[string][]byte{
		"dataConfigJSON": dataConfigJSON,
		"dataConfigYAML": dataConfigYAML,

		"nvme_raidData": nvme_raidData,
	} {
		require.NotNil(t, data, name)
	}
}

func TestNvme_Raid_ConfigurationSerialize(t *testing.T) {
	module.TestConfigurationSerialize(t, &Nvme_Raid{}, dataConfigJSON, dataConfigYAML)
}

func TestNvme_Raid_Init(t *testing.T) {
	tests := map[string]struct {
		config   Config
		wantFail bool
	}{
		"fails if 'ndsudo' not found": {
			wantFail: true,
			config:   New().Config,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			stor := New()

			if test.wantFail {
				assert.Error(t, stor.Init())
			} else {
				assert.NoError(t, stor.Init())
			}
		})
	}
}

func TestNvme_Raid_Cleanup(t *testing.T) {
	tests := map[string]struct {
		prepare func() *Nvme_Raid
	}{
		"not initialized exec": {
			prepare: func() *Nvme_Raid {
				return New()
			},
		},
		"after check": {
			prepare: func() *Nvme_Raid {
				stor := New()
				stor.exec = prepareMockNvme_RaidOK()
				_ = stor.Check()
				return stor
			},
		},
		"after collect": {
			prepare: func() *Nvme_Raid {
				stor := New()
				stor.exec = prepareMockNvme_RaidOK()
				_ = stor.Collect()
				return stor
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			stor := test.prepare()

			assert.NotPanics(t, stor.Cleanup)
		})
	}
}

func TestNvme_Raid_Charts(t *testing.T) {
	assert.NotNil(t, New().Charts())
}

func TestNvme_Raid_Check(t *testing.T) {
	tests := map[string]struct {
		prepareMock func() *mockNvme_RaidExec
		wantFail    bool
	}{
		"success Nvme_Raid Raid": {
			wantFail:    false,
			prepareMock: prepareMockNvme_RaidOK,
		},
		"err on exec": {
			wantFail:    true,
			prepareMock: prepareMockErr,
		},
		"unexpected response": {
			wantFail:    true,
			prepareMock: prepareMockUnexpectedResponse,
		},
		"empty response": {
			wantFail:    true,
			prepareMock: prepareMockEmptyResponse,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			stor := New()
			mock := test.prepareMock()
			stor.exec = mock

			if test.wantFail {
				assert.Error(t, stor.Check())
			} else {
				assert.NoError(t, stor.Check())
			}
		})
	}
}

func TestNvme_Raid_Collect(t *testing.T) {
	tests := map[string]struct {
		prepareMock func() *mockNvme_RaidExec
		wantMetrics map[string]int64
		wantCharts  int
	}{
		"success Nvme_Raid Raid": {
			prepareMock: prepareMockNvme_RaidOK,
			wantCharts:  len(raidDataChartsTmpl) * 2,
			wantMetrics: map[string]int64{},
		},
		"err on exec": {
			prepareMock: prepareMockErr,
			wantMetrics: nil,
		},
		"unexpected response": {
			prepareMock: prepareMockUnexpectedResponse,
			wantMetrics: nil,
		},
		"empty response": {
			prepareMock: prepareMockEmptyResponse,
			wantMetrics: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			stor := New()
			mock := test.prepareMock()
			stor.exec = mock

			mx := stor.Collect()

			assert.Equal(t, test.wantMetrics, mx)
			assert.Len(t, *stor.Charts(), test.wantCharts)
			testMetricsHasAllChartsDims(t, stor, mx)
		})
	}
}

func prepareMockNvme_RaidOK() *mockNvme_RaidExec {
	return &mockNvme_RaidExec{
		raid_dataInfoData: nvme_raidData,
	}
}

func prepareMockErr() *mockNvme_RaidExec {
	return &mockNvme_RaidExec{
		errOnInfo: true,
	}
}

func prepareMockUnexpectedResponse() *mockNvme_RaidExec {
	resp := []byte(`
Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Nulla malesuada erat id magna mattis, eu viverra tellus rhoncus.
Fusce et felis pulvinar, posuere sem non, porttitor eros.
`)
	return &mockNvme_RaidExec{
		raid_dataInfoData: resp,
	}
}

func prepareMockEmptyResponse() *mockNvme_RaidExec {
	return &mockNvme_RaidExec{}
}

type mockNvme_RaidExec struct {
	errOnInfo         bool
	raid_dataInfoData []byte
}

func (m *mockNvme_RaidExec) nvme_raidInfo() ([]byte, error) {
	if m.errOnInfo {
		return nil, errors.New("mock.raid_data() error")
	}
	return m.raid_dataInfoData, nil
}

func testMetricsHasAllChartsDims(t *testing.T, stor *Nvme_Raid, mx map[string]int64) {
	for _, chart := range *stor.Charts() {
		if chart.Obsolete {
			continue
		}
		for _, dim := range chart.Dims {
			_, ok := mx[dim.ID]
			assert.Truef(t, ok, "collected metrics has no data for dim '%s' chart '%s'", dim.ID, chart.ID)
		}
		for _, v := range chart.Vars {
			_, ok := mx[v.ID]
			assert.Truef(t, ok, "collected metrics has no data for var '%s' chart '%s'", v.ID, chart.ID)
		}
	}
}
