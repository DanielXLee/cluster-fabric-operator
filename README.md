
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*
<p align="center">
  <a href="https://github.com/DanielXLee/cluster-fabric-operator">
    <img src="https://github.com/DanielXLee/cluster-fabric-operator/workflows/CI%20Pipeline/badge.svg" alt="Github CI">
  </a>
  <a href="https://goreportcard.com/report/github.com/DanielXLee/cluster-fabric-operator">
    <img src="https://goreportcard.com/badge/github.com/DanielXLee/cluster-fabric-operator" alt="GoReportCard">
  </a>
  <a href="http://golang.org">
    <img src="https://img.shields.io/github/go-mod/go-version/DanielXLee/cluster-fabric-operator" alt="GitHub go.mod Go version (subdirectory of monorepo)">
  </a>
  <a href="http://golang.org">
    <img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg" alt="made-with-Go">
  </a>
  <a href="https://quay.io/repository/opstree/cluster-fabric-operator">
    <img src="https://img.shields.io/badge/container-ready-green" alt="Docker">
  </a>
  <a href="https://github.com/DanielXLee/cluster-fabric-operator/master/LICENSE">
    <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License">
  </a>
</p>
- [Cluster Fabric Operator](#cluster-fabric-operator)
  - [Architecture](#architecture)
    - [Purpose](#purpose)
    - [Supported Features](#supported-features)
    - [Getting Started](#getting-started)
    - [Example](#example)
    - [Prerequisites](#prerequisites)
    - [Quickstart](#quickstart)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Cluster Fabric Operator

A Golang based fabric operator that will make/oversee Submariner components on top of the Kubernetes.

## Architecture


### Purpose

The purpose of creating this operator was to provide an easy and production-grade setup of Submariner components on Kubernetes. It doesn't care if you have a plain on-prem Kubernetes or cloud-based.

### Supported Features

Here the features which are supported by this operator:-

- Deploy submariner broker
- Join managed cluster to broker
- Check k8s server version
- Support cloud prepare (AWS, GCE)
- Support components enable/disable

### Getting Started

### Example

The configuration of Fabric setup should be described in Fabric CRD. You will find all the examples manifests in [example](./connfig/samples) folder.

### Prerequisites

Fabric operator requires a Kubernetes cluster of version `>=1.7.0`. If you have just started with Operators, its highly recommended to use latest version of Kubernetes.

### Quickstart

The setup can be done by using `kustomize`.

```bash
$ git clone https://github.com/DanielXLee/cluster-fabric-operator.git
```

```bash
$ cd cluster-fabric-operator
$ make deploy
```
