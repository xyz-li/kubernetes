/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package containermap

import (
	"fmt"
	"sync"
)

// ContainerMapWithLock add a RWMutex to guard the map ContainerMap.
// ContainerMap maps (containerID)->(*v1.Pod, *v1.Container)
type ContainerMapWithLock struct {
	sync.RWMutex
	ContainerMap
}

func NewContainerMapWithLock() *ContainerMapWithLock {
	return &ContainerMapWithLock{
		ContainerMap: make(ContainerMap),
	}
}

// Add adds a mapping of (containerID)->(podUID, containerName) to the ContainerMap
func (cm *ContainerMapWithLock) Add(podUID, containerName, containerID string) {
	cm.Lock()
	cm.ContainerMap[containerID] = struct {
		podUID        string
		containerName string
	}{podUID, containerName}
	cm.Unlock()
}

// RemoveByContainerID removes a mapping of (containerID)->(podUID, containerName) from the ContainerMap
func (cm *ContainerMapWithLock) RemoveByContainerID(containerID string) {
	cm.Lock()
	delete(cm.ContainerMap, containerID)
	cm.Unlock()
}

// RemoveByContainerRef removes a mapping of (containerID)->(podUID, containerName) from the ContainerMap
func (cm *ContainerMapWithLock) RemoveByContainerRef(podUID, containerName string) {
	cm.Lock()
	containerID, err := cm.GetContainerID(podUID, containerName)
	if err == nil {
		cm.RemoveByContainerID(containerID)
	}
	cm.Unlock()
}

// GetContainerID retrieves a ContainerID from the ContainerMap
func (cm ContainerMapWithLock) GetContainerID(podUID, containerName string) (string, error) {
	cm.RLock()
	defer cm.RUnlock()
	for key, val := range cm.ContainerMap {
		if val.podUID == podUID && val.containerName == containerName {
			return key, nil
		}
	}
	return "", fmt.Errorf("container %s not in ContainerMap for pod %s", containerName, podUID)
}

// GetContainerRef retrieves a (podUID, containerName) pair from the ContainerMap
func (cm *ContainerMapWithLock) GetContainerRef(containerID string) (string, string, error) {
	cm.RLock()
	defer cm.RUnlock()
	if _, exists := cm.ContainerMap[containerID]; !exists {
		return "", "", fmt.Errorf("containerID %s not in ContainerMap", containerID)
	}
	return cm.ContainerMap[containerID].podUID, cm.ContainerMap[containerID].containerName, nil
}

// Visit invoke visitor function to walks all of the entries in the container map
func (cm *ContainerMapWithLock) Visit(visitor func(podUID, containerName, containerID string)) {
	cm.Lock()
	defer cm.Unlock()
	for k, v := range cm.ContainerMap {
		visitor(v.podUID, v.containerName, k)
	}
}
