<!---
  SPDX-FileCopyrightText: (C) 2024 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Metrics Library for time measurement

We have introduced two metrics:

1. **event_timestamp**: This metric records the timestamp of any event occurring.
    For E.g. We have used this metric to record timestamp of the following
    events in app-deployment-manager :
    1. Start of Deployment custom resource creation.
    2. Start of Gitrepo Resource creation.
    3. Start of DeploymentCluster custom resource creation.
    4. Status change of Deployment resource to Running State.
    5. Status change of DeploymentCluster resource to Running State.

2. **time_difference_between_events**: This metric records the time difference
between 2 specified events.
    For E.g. we have used this metric to record the following timings:
    1. Time between deployment creation and deployment going to running state
    2. Time between deployment creation and deploymentcluster going to running state
    3. Time between deployment creation and Gitrepo creation
    4. Time between deployment creation and DeploymentCluster creation

# **Notes**

1. If the Time between deployment creation and deployment going to running
state keeps increasing significantly with every deployment, then its gives an
indication that there is some component which is causing the bottleneck.
2. We can look at Time between deployment creation and Gitrepo creation and if
that difference is high, then we should look at gitea component to see what the
issue is.
3. If gitrepo creation time is not significant, then we look at the Time between
deployment creation and DeploymentCluster creation. If this difference is high,
then it means that the cluster is taking time to respond.
4. If deploymentcluster creation is not significant, then we look into the Time
between deployment creation and deploymentcluster going to running state. This
indirectly points us to the bundledeployment of the application. This means, we
need to now look at the edge node and see why the pods are not coming up.
