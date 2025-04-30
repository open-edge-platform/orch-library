// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"time"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/open-edge-platform/app-orch-deployment/app-deployment-manager/api/v1beta1"
)

var log = dazl.GetPackageLogger()

var (
	// Custom collector
	MeasurementReg = prometheus.NewRegistry()

	EventTimestampGuage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "event_timestamp",
			Help: "Timestamp of occurunce of event",
		},
		[]string{"projectID", "deploymentID", "displayName", "part", "event"},
	)

	EventTimeDifferenceGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "time_difference_between_events",
			Help: "Time difference between first and last timestamp",
		},
		[]string{"projectID", "deploymentID", "displayName", "firstPart", "firstEvent",
        "lastPart", "lastEvent"},
	)

	// Map to store timestamps for each deployment
	Timestamps = make(map[string]map[string]float64)
)

func RecordTimestamp(projectID, deploymentID, displayName, part, event string) {
	log.Infof("record timestamp %s %s %s %s", projectID, deploymentID, part, event)
	timestamp := float64(time.Now().Unix())
	key := part + "_" + event

	// Initialize the map for the deployment if it doesn't exist
	if _, exists := Timestamps[deploymentID]; !exists {
		Timestamps[deploymentID] = make(map[string]float64)
	}

	// Write timestamp for running state only if entry doesn't exist
	if part == string(v1beta1.Running) {
		if _, exists := Timestamps[deploymentID][key]; !exists {
			Timestamps[deploymentID][key] = timestamp
			EventTimestampGuage.WithLabelValues(projectID, deploymentID, displayName, part, event).Set(timestamp)
			log.Infof("write timestamp for changing to running state  %s, %s, %s", projectID, deploymentID, event)
		}
	} else {
		Timestamps[deploymentID][key] = timestamp
		EventTimestampGuage.WithLabelValues(projectID, deploymentID, displayName, part, event).Set(timestamp)
	}
}

func DeleteTimestampMetrics(projectID, deploymentID string) {
	log.Infof("delete timestamp %s %s", projectID, deploymentID)
	delete(Timestamps, deploymentID)
}

func CalculateTimeDifference(projectID, deploymentID, displayName, firstPart, firstEvent, lastPart, lastEvent string) {
	firstKey := firstPart + "_" + firstEvent
	lastKey := lastPart + "_" + lastEvent

	firstTimestamp, firstExists := Timestamps[deploymentID][firstKey]
	lastTimestamp, lastExists := Timestamps[deploymentID][lastKey]

	if firstExists && lastExists {
		timeDifference := lastTimestamp - firstTimestamp
		EventTimeDifferenceGauge.WithLabelValues(projectID, deploymentID, displayName, firstPart, firstEvent, lastPart, lastEvent).Set(timeDifference)
	}
}
func init() {
	// Register custom metrics with prometheus registry
	log.Infof("metrics server init \n")
	MeasurementReg.MustRegister(EventTimestampGuage, EventTimeDifferenceGauge)
}
