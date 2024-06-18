# -*- coding: utf-8 -*-
# Description: example netdata python.d module
# Author: Your name (your github login)
# SPDX-License-Identifier: GPL-3.0-or-later

import subprocess
import json
import logging

from bases.FrameworkServices.SimpleService import SimpleService  # type: ignore

# Set up logging
logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
stream_handler = logging.StreamHandler()
stream_handler.setFormatter(formatter)
logger.addHandler(stream_handler)

priority = 90000


class Service(SimpleService):
    def __init__(self, configuration=None, name=None):
        SimpleService.__init__(self, configuration=configuration, name=name)
        self.order = []
        self.definitions = {}
        self.raid_states = {}

    def fetch_raid_info(self):
        try:
            output = subprocess.check_output(["eraraid", "show", "-f", "json", "-e"])
        except FileNotFoundError:
            try:
                output = subprocess.check_output(["xiraid", "show", "-f", "json", "-e"])
            except FileNotFoundError:
                logger.error("Neither eraraid nor xiraid command found.")
                return {}

        resp = json.loads(output.decode("utf-8"))
        return resp

    def collect_raid_info(self):
        raid_data = self.fetch_raid_info()
        if raid_data:
            # print(raid_data)  # Debugging: Print fetched RAID data
            raid_states = {}
            for raid_name, raid in raid_data.items():
                raid_states[raid_name] = {state.lower(): 0 for state in [
                    "online", "initialized", "initing", "degraded", "reconstructing",
                    "offline", "need_recon", "need_init", "read_only", "unrecovered",
                    "none", "restriping", "need_resize", "need_restripe"
                ]}
                for state in raid.get("state", []):
                    raid_states[raid_name][state.lower()] = 1
            return raid_states
        else:
            return None

    def get_data(self):
        raid_states = self.collect_raid_info()
        if raid_states:
            data = {}
            for raid_name, states in raid_states.items():
                for state, value in states.items():
                    dimension_id = f"raid_{raid_name}_state_{state.replace(' ', '_')}"
                    data[dimension_id] = value
            return data
        else:
            return {}

    def create_chart(self, raid_name, states):
        chart_id = f"raid_{raid_name}_state"
        chart_title = f"RAID State - {raid_name}"
        options = ["", chart_title, "status", "raid state", "storcli.raid_state", "line"]
        lines = [[state] for state in states]
        self.order.append(chart_id)
        self.definitions[chart_id] = {
            "options": options,
            "lines": lines
        }

    def setup_charts(self, raid_states):
        for raid_name, states in raid_states.items():
            active_states = [state for state, value in states.items() if value == 1]
            if len(active_states) > 3:
                active_states = active_states[:3]
            self.create_chart(raid_name, active_states)

    def check(self):
        raid_states = self.collect_raid_info()
        if raid_states:
            self.setup_charts(raid_states)
            return True
        return False


if __name__ == "__main__":
    service = Service()
    service.run()
