// Code generated by mockery v1.0.0

// Regenerate this file using `make store-mocks`.

package mocks

import mock "github.com/stretchr/testify/mock"
import model "github.com/SoulDemon/mattermostp/model"
import store "github.com/SoulDemon/mattermostp/store"

// ClusterDiscoveryStore is an autogenerated mock type for the ClusterDiscoveryStore type
type ClusterDiscoveryStore struct {
	mock.Mock
}

// Cleanup provides a mock function with given fields:
func (_m *ClusterDiscoveryStore) Cleanup() store.StoreChannel {
	ret := _m.Called()

	var r0 store.StoreChannel
	if rf, ok := ret.Get(0).(func() store.StoreChannel); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(store.StoreChannel)
		}
	}

	return r0
}

// Delete provides a mock function with given fields: discovery
func (_m *ClusterDiscoveryStore) Delete(discovery *model.ClusterDiscovery) store.StoreChannel {
	ret := _m.Called(discovery)

	var r0 store.StoreChannel
	if rf, ok := ret.Get(0).(func(*model.ClusterDiscovery) store.StoreChannel); ok {
		r0 = rf(discovery)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(store.StoreChannel)
		}
	}

	return r0
}

// Exists provides a mock function with given fields: discovery
func (_m *ClusterDiscoveryStore) Exists(discovery *model.ClusterDiscovery) store.StoreChannel {
	ret := _m.Called(discovery)

	var r0 store.StoreChannel
	if rf, ok := ret.Get(0).(func(*model.ClusterDiscovery) store.StoreChannel); ok {
		r0 = rf(discovery)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(store.StoreChannel)
		}
	}

	return r0
}

// GetAll provides a mock function with given fields: discoveryType, clusterName
func (_m *ClusterDiscoveryStore) GetAll(discoveryType string, clusterName string) store.StoreChannel {
	ret := _m.Called(discoveryType, clusterName)

	var r0 store.StoreChannel
	if rf, ok := ret.Get(0).(func(string, string) store.StoreChannel); ok {
		r0 = rf(discoveryType, clusterName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(store.StoreChannel)
		}
	}

	return r0
}

// Save provides a mock function with given fields: discovery
func (_m *ClusterDiscoveryStore) Save(discovery *model.ClusterDiscovery) store.StoreChannel {
	ret := _m.Called(discovery)

	var r0 store.StoreChannel
	if rf, ok := ret.Get(0).(func(*model.ClusterDiscovery) store.StoreChannel); ok {
		r0 = rf(discovery)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(store.StoreChannel)
		}
	}

	return r0
}

// SetLastPingAt provides a mock function with given fields: discovery
func (_m *ClusterDiscoveryStore) SetLastPingAt(discovery *model.ClusterDiscovery) store.StoreChannel {
	ret := _m.Called(discovery)

	var r0 store.StoreChannel
	if rf, ok := ret.Get(0).(func(*model.ClusterDiscovery) store.StoreChannel); ok {
		r0 = rf(discovery)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(store.StoreChannel)
		}
	}

	return r0
}
