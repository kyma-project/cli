// Code generated by mockery v2.38.0. DO NOT EDIT.

package mocks

import (
	comparch "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	mock "github.com/stretchr/testify/mock"

	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm"
)

// OciRepoAccess is an autogenerated mock type for the OciRepoAccess type
type OciRepoAccess struct {
	mock.Mock
}

// ComponentVersionExists provides a mock function with given fields: archive, repo
func (_m *OciRepoAccess) ComponentVersionExists(archive *comparch.ComponentArchive, repo ocm.Repository) (bool, error) {
	ret := _m.Called(archive, repo)

	if len(ret) == 0 {
		panic("no return value specified for ComponentVersionExists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(*comparch.ComponentArchive, ocm.Repository) (bool, error)); ok {
		return rf(archive, repo)
	}
	if rf, ok := ret.Get(0).(func(*comparch.ComponentArchive, ocm.Repository) bool); ok {
		r0 = rf(archive, repo)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(*comparch.ComponentArchive, ocm.Repository) error); ok {
		r1 = rf(archive, repo)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DescriptorResourcesAreEquivalent provides a mock function with given fields: archive, remoteVersion
func (_m *OciRepoAccess) DescriptorResourcesAreEquivalent(archive *comparch.ComponentArchive, remoteVersion ocm.ComponentVersionAccess) bool {
	ret := _m.Called(archive, remoteVersion)

	if len(ret) == 0 {
		panic("no return value specified for DescriptorResourcesAreEquivalent")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(*comparch.ComponentArchive, ocm.ComponentVersionAccess) bool); ok {
		r0 = rf(archive, remoteVersion)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GetComponentVersion provides a mock function with given fields: archive, repo
func (_m *OciRepoAccess) GetComponentVersion(archive *comparch.ComponentArchive, repo ocm.Repository) (ocm.ComponentVersionAccess, error) {
	ret := _m.Called(archive, repo)

	if len(ret) == 0 {
		panic("no return value specified for GetComponentVersion")
	}

	var r0 ocm.ComponentVersionAccess
	var r1 error
	if rf, ok := ret.Get(0).(func(*comparch.ComponentArchive, ocm.Repository) (ocm.ComponentVersionAccess, error)); ok {
		return rf(archive, repo)
	}
	if rf, ok := ret.Get(0).(func(*comparch.ComponentArchive, ocm.Repository) ocm.ComponentVersionAccess); ok {
		r0 = rf(archive, repo)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ocm.ComponentVersionAccess)
		}
	}

	if rf, ok := ret.Get(1).(func(*comparch.ComponentArchive, ocm.Repository) error); ok {
		r1 = rf(archive, repo)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PushComponentVersion provides a mock function with given fields: archive, repository, overwrite
func (_m *OciRepoAccess) PushComponentVersion(archive *comparch.ComponentArchive, repository ocm.Repository, overwrite bool) error {
	ret := _m.Called(archive, repository, overwrite)

	if len(ret) == 0 {
		panic("no return value specified for PushComponentVersion")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*comparch.ComponentArchive, ocm.Repository, bool) error); ok {
		r0 = rf(archive, repository, overwrite)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewOciRepoAccess creates a new instance of OciRepoAccess. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOciRepoAccess(t interface {
	mock.TestingT
	Cleanup(func())
}) *OciRepoAccess {
	mock := &OciRepoAccess{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
