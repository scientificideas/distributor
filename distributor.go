/*
Copyright Scientific Ideas 2022. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/scientificideas/distributor/pinger"
	"github.com/scientificideas/distributor/storage"
	"github.com/serialx/hashring"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultPingTimeout  = 800 * time.Millisecond // default time to wait for service response
	defaultPollInterval = 500 * time.Millisecond // default service ping interval
)

// Distributor manages SeviceCache, checks services liveness, and updates services-to-work-units matching table.
type Distributor struct {
	Storage               storage.Storage
	ringMembers           string
	serviceNamespace      string
	distributionNamespace string
	p                     pinger.Pinger
	serviceCache          ServiceCache
	workUnitsCache        WorkUnitsCache
	transport             *Transport
}

// Transport configures network parameters of Distributor.
type Transport struct {
	PingTimeout  time.Duration
	PollInterval time.Duration
}

// ServiceCache manages a cache of active services for Distributor.
type ServiceCache struct {
	mu       sync.RWMutex // mutex for services map
	services map[string]struct{}
}

// WorkUnitsCache manages a cache of active services for Distributor.
type WorkUnitsCache struct {
	mu        sync.RWMutex // mutex for work units map
	workunits map[string]struct{}
}

// NewDistributor creates a Distributor instance.
// ringMembers ang serviceNamespace are the keys in storage where work units list and services list are stored respectively.
func NewDistributor(distributionNamespace, ringMembers, serviceNamespace string, p pinger.Pinger, opts ...Option) (*Distributor, error) {
	d := &Distributor{p: p}
	for _, opt := range opts {
		if err := opt(d); err != nil {
			return nil, err
		}
	}
	if d.transport.PingTimeout == 0 {
		d.transport.PingTimeout = defaultPingTimeout
	}
	if d.transport.PollInterval == 0 {
		d.transport.PollInterval = defaultPollInterval
	}
	if distributionNamespace == "" {
		return nil, errors.New("got empty work distribution namespace")
	}
	d.distributionNamespace = distributionNamespace
	if ringMembers == "" {
		return nil, errors.New("ring members list key is empty")
	}
	d.ringMembers = ringMembers
	if serviceNamespace == "" {
		return nil, errors.New("service namespace is not set (-services-namespaces= or SERVICES_NAMESPACES)")
	}
	d.serviceNamespace = serviceNamespace
	d.serviceCache.services = make(map[string]struct{})
	d.workUnitsCache.workunits = make(map[string]struct{})
	urls, err := d.Services()
	if err != nil {
		return nil, err
	}
	if err = d.p.Init(urls...); err != nil {
		return nil, err
	}
	return d, nil
}

// PutToMatchingTable creates hash table in storage where each service has its own range of ring hash members.
func (d *Distributor) PutToMatchingTable(services, ringMembers []string) error {
	ring := hashring.NewWithWeights(nil)
	for _, member := range ringMembers {
		ring = ring.AddWeightedNode(member, 50)
	}

	// find matching ring members for every item
	var matchingTable = make(map[string]interface{})
	//var listOfItems string
	itemsCountPerMember := len(ringMembers) / len(services)
	if itemsCountPerMember < 1 {
		itemsCountPerMember = 1
	}

	sort.Strings(services) // sort services list for consistent results
	for i, service := range services {
		if i == len(services)-1 && len(ringMembers) > len(services) { // if it's the last member and we have more free buckets than itemsCountPerMember,
			itemsCountPerMember = itemsCountPerMember + len(ringMembers)%len(services) // give to this member itemsCountPerMember + len(ringMembers) % len(services) buckets
		}
		matchingMembers, ok := ring.GetNodes(service, itemsCountPerMember)
		if !ok {
			logrus.Debugf("failed to find matching hash ring member for item %s, skipping", service)
		}
		// rm member from hash ring to avoid duplication: one ring member can't be associated with two services
		for _, member := range matchingMembers {
			ring = ring.RemoveNode(member)
		}
		matchingTable[service] = strings.Join(matchingMembers, ",")
	}

	logrus.Debugf("new matching table: %v", matchingTable)
	return d.Storage.SetMap(d.distributionNamespace, matchingTable)
}

func (d *Distributor) balance() error {
	ringMembers, err := d.RingMembers()
	if err != nil {
		return err
	}
	if len(ringMembers) == 0 {
		return fmt.Errorf("units of work not found in storage")
	}
	services, err := d.Services()
	if err != nil {
		return err
	}
	if len(services) == 0 {
		return fmt.Errorf("no one service responded")
	}
	return d.PutToMatchingTable(services, ringMembers)
}

// Services returns all services registered in storage.
func (d *Distributor) Services() ([]string, error) {
	return d.Storage.GetList(d.serviceNamespace)
}

// RingMembers returns all ring members (work units) registered in storage.
func (d *Distributor) RingMembers() ([]string, error) {
	return d.Storage.GetList(d.ringMembers)
}

func (s *ServiceCache) add(service string) {
	s.mu.Lock()
	s.services[service] = struct{}{}
	s.mu.Unlock()
}

func (s *ServiceCache) del(service string) {
	s.mu.Lock()
	delete(s.services, service)
	s.mu.Unlock()
}

func (s *ServiceCache) exist(service string) bool {
	var serviceFound bool
	s.mu.RLock()
	_, serviceFound = s.services[service]
	s.mu.RUnlock()
	return serviceFound
}

func (s *ServiceCache) all() []string {
	var services []string
	s.mu.RLock()
	for service := range s.services {
		services = append(services, service)
	}
	s.mu.RUnlock()
	return services
}

func (w *WorkUnitsCache) add(workunit string) {
	w.mu.Lock()
	w.workunits[workunit] = struct{}{}
	w.mu.Unlock()
}

func (s *WorkUnitsCache) del(workunit string) {
	s.mu.Lock()
	delete(s.workunits, workunit)
	s.mu.Unlock()
}

func (s *WorkUnitsCache) exist(workunit string) bool {
	var workUnitFound bool
	s.mu.RLock()
	_, workUnitFound = s.workunits[workunit]
	s.mu.RUnlock()
	return workUnitFound
}

func (s *WorkUnitsCache) all() []string {
	var workUnits []string
	s.mu.RLock()
	for workUnit := range s.workunits {
		workUnits = append(workUnits, workUnit)
	}
	s.mu.RUnlock()
	return workUnits
}

func includes(source []string, item string) bool {
	for _, s := range source {
		if s == item {
			return true
		}
	}
	return false
}

// LivenessCheck checks current active services by ping them, checks storage for new services and rebalance work units
// if new units of work (ring members) appear in the storage, they will be distributed among services in the balance() method call.
func (d *Distributor) LivenessCheck() error {
	servicesFromStorage, err := d.Services()
	if err != nil {
		return err
	}
	workunitsFromStorage, err := d.RingMembers()
	if err != nil {
		return err
	}

	// check that all services from cache still exist in storage
	allCachedServices := d.serviceCache.all()
	for _, serviceFromCache := range allCachedServices {
		// service deleted from storage, let's rebalance matching table in storage and del redundant service from cache
		if !includes(servicesFromStorage, serviceFromCache) {
			logrus.Debugf("%s service deleted from the storage, rebalance and del redundant item from the cache", serviceFromCache)
			// del from local cache
			logrus.Debugf("delete service %s from the cache of the %s namespace", serviceFromCache, d.serviceNamespace)
			d.serviceCache.del(serviceFromCache)
			// del from storage hash table
			logrus.Debugf("delete service %s from the storage matching table", serviceFromCache)
			if err = d.Storage.DelFromMap(d.distributionNamespace, serviceFromCache); err != nil {
				return err
			}
			// rebalance
			if err := d.balance(); err != nil {
				return err
			}
		}
	}

	for _, service := range servicesFromStorage {
		// rebalance if this service doesn't exist in distributor cache
		if !d.serviceCache.exist(service) {
			logrus.Debugf("no service %s in the cache of the %s namespace, rebalance", service, d.serviceNamespace)
			if err := d.balance(); err != nil {
				logrus.Warn(err)
			} else {
				logrus.Debugf("add service %s to the cache", service)
				d.serviceCache.add(service)
			}
		}

		// rebalance if service doesn't respond correctly (timing,service error network errors, service fault)
		ctx, _ := context.WithTimeout(context.Background(), d.transport.PingTimeout)
		if err := d.p.Ping(ctx, service); err != nil {
			logrus.Warnf("ping %s service error: %s", service, err)
			// rm faulty service from cache
			logrus.Warnf("delete %s service from the local cache of the %s namespace", service, d.serviceNamespace)
			d.serviceCache.del(service)

			// del from storage services list
			logrus.Warnf("delete %s service from the storage services list", service)
			if err = d.Storage.DelFromList(d.serviceNamespace, service); err != nil {
				return err
			}
			// del from storage hash table
			logrus.Warnf("delete %s service from storage matching table", service)
			if err = d.Storage.DelFromMap(d.distributionNamespace, service); err != nil {
				return err
			}
			// rebalance
			if err := d.balance(); err != nil {
				return err
			}
		}
	}

	// check work units
	allCachedWorkUnits := d.workUnitsCache.all()
	for _, workUnitFromCache := range allCachedWorkUnits {
		// workunit deleted from storage, let's rebalance matching table in storage and del redundant workunit from cache
		if !includes(workunitsFromStorage, workUnitFromCache) {
			logrus.Debugf("%s work unit deleted from the storage, rebalance and del redundant item from the cache of the %s namespace", workUnitFromCache, d.serviceNamespace)
			// del from local cache
			logrus.Debugf("delete work unit %s from the cache of the %s namespace", workUnitFromCache, d.serviceNamespace)
			d.workUnitsCache.del(workUnitFromCache)
			// del from storage hash table
			logrus.Debugf("delete work unit %s from the storage matching table", workUnitFromCache)
			if err = d.Storage.DelFromMap(d.distributionNamespace, workUnitFromCache); err != nil {
				return err
			}
			// rebalance
			if err := d.balance(); err != nil {
				return err
			}
		}
	}

	for _, workUnitFromStorage := range workunitsFromStorage {
		// rebalance if this work unit doesn't exist in distributor cache
		if !d.workUnitsCache.exist(workUnitFromStorage) {
			logrus.Debugf("no work unit %s in the cache of the %s namespace, rebalance", workUnitFromStorage, d.serviceNamespace)
			if err := d.balance(); err != nil {
				logrus.Warn(err)
			} else {
				logrus.Debugf("add work unit %s to the cache of the %s namespace", workUnitFromStorage, d.serviceNamespace)
				d.workUnitsCache.add(workUnitFromStorage)
			}
		}
	}

	return nil
}

// Run performs a liveness check and work units distribution at the interval specified in 'pollInterval' arg.
func (d *Distributor) Run(errorsChan chan error) {
	t := time.NewTicker(d.transport.PollInterval)
	for {
		<-t.C
		if err := d.LivenessCheck(); err != nil {
			errorsChan <- err
		}
	}
}
