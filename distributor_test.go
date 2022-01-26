/*
Copyright LLC Newity. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/scientificideas/distributor/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

var TestTable = map[string]TestData{
	"TestPutToMatchingTable": {"testMatchingTablKey", "testWorkUnitsKey", "testServicesNamespace1,testServicesNamespace2",
		[]string{"service1", "service2", "service3"},
		[]string{"work1", "work2", "work3"},
		[]ExpectedMatchingTable{
			{"service1", "work2"},
			{"service2", "work1"},
			{"service3", "work3"},
		},
	},
	"TestLivenessCheck": {"testMatchingTablKey", "testWorkUnitsKey", "testServicesNamespace1,testServicesNamespace2",
		[]string{"service1", "service2", "service3", "badService1", "badService2", "badService3"},
		[]string{"work1", "work2", "work3"},
		[]ExpectedMatchingTable{
			{"service1", "work2"},
			{"service2", "work1"},
			{"service3", "work3"},
		},
	},
}

type TestData struct {
	DistributionNamespace string
	RingMembersKey        string
	ServicesListsKeys     string
	Services              []string
	WorkUnits             []string
	resultMatchingTable   []ExpectedMatchingTable
}

type ExpectedMatchingTable struct {
	Service string
	Work    string
}

func TestPutToMatchingTable(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	// create Distributor with mocked Storage and Pinger
	distributor, err := CreateDistributor(TestTable["TestPutToMatchingTable"])
	assert.NoError(t, err)

	servicesFromStorage, workunitsFromStorage, err := GetServicesAndRingMembers(distributor)
	assert.NoError(t, err)

	// create matching table
	assert.NoError(t, distributor.PutToMatchingTable(servicesFromStorage, workunitsFromStorage))

	// check that the distribution matches the expected
	for _, service := range servicesFromStorage {
		foundWork, err := distributor.Storage.GetMapField(TestTable["TestPutToMatchingTable"].DistributionNamespace, service)
		assert.NoError(t, err)
		foundWorkString := strings.Join(foundWork, ",")
		for _, expected := range TestTable["TestPutToMatchingTable"].resultMatchingTable {
			if foundWorkString == expected.Work {
				assert.Equal(t, expected.Service, service)
			}
		}
	}
}

func TestLivenessCheck(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	// create Distributor with mocked Storage and Pinger
	distributor, err := CreateDistributor(TestTable["TestPutToMatchingTable"])
	assert.NoError(t, err)

	servicesFromStorage, workunitsFromStorage, err := GetServicesAndRingMembers(distributor)
	assert.NoError(t, err)
	// create matching table
	assert.NoError(t, distributor.PutToMatchingTable(servicesFromStorage, workunitsFromStorage))

	// check nodes liveness
	distributor.LivenessCheck()

	// create matching table
	assert.NoError(t, distributor.PutToMatchingTable(servicesFromStorage, workunitsFromStorage))

	// check that the distribution matches the expected and the faulty nodes are not included into this distribution
	for _, service := range servicesFromStorage {
		foundWork, err := distributor.Storage.GetMapField(TestTable["TestLivenessCheck"].DistributionNamespace, service)
		assert.NoError(t, err)
		foundWorkString := strings.Join(foundWork, ",")
		for _, expected := range TestTable["TestLivenessCheck"].resultMatchingTable {
			if foundWorkString == expected.Work {
				assert.Equal(t, expected.Service, service)
			}
		}
	}
}

func CreateDistributor(testData TestData) (*Distributor, error) {
	mockStorage := &mocks.MockStorage{
		Lists:     make(map[string][]string),
		HashTable: make(map[string]string),
	}
	mockStorage.Lists[testData.ServicesListsKeys] = testData.Services
	mockStorage.Lists[testData.RingMembersKey] = testData.WorkUnits

	return NewDistributor(testData.DistributionNamespace, testData.RingMembersKey, testData.ServicesListsKeys,
		mocks.NewMockPinger(), WithStorage(mockStorage), WithTransport(&Transport{
			PingTimeout:  time.Millisecond,
			PollInterval: time.Millisecond,
		}))
}

func GetServicesAndRingMembers(distributor *Distributor) ([]string, []string, error) {
	servicesFromStorage, err := distributor.Services()
	if err != nil {
		return nil, nil, err
	}
	workunitsFromStorage, err := distributor.RingMembers()
	if err != nil {
		return nil, nil, err
	}
	return servicesFromStorage, workunitsFromStorage, nil
}
