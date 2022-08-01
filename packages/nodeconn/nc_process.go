// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import "time"

func (nc *nodeConn) run() {
	for {
		err := nc.mqttClient.Connect(nc.ctx)
		if err == nil {
			break
		}
		nc.log.Warnf("Unable to connect to Hornet MQTT: %v", err)
		time.Sleep(1 * time.Second)
	}
	milestones, subInfo := nc.mqttClient.ConfirmedMilestones()
	if subInfo.Error() != nil {
		nc.log.Panicf("Error subscribing: %v", subInfo.Error())
	}
	for {
		select {
		case m, ok := <-milestones:
			nc.log.Debugf("Milestone received, index=%v, timestamp=%v", m.Index, m.Timestamp)
			if ok {
				nc.metrics.GetInMilestone().CountLastMessage(m)
				nc.milestones.Trigger(m)
			}
		case <-nc.ctx.Done():
			return
		}
	}
}
