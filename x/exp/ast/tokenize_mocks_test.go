// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package ast

import (
	"sync"
)

// Ensure, that readerMock does implement reader.
// If this is not the case, regenerate this file with moq.
var _ reader = &readerMock{}

// readerMock is a mock implementation of reader.
//
//	func TestSomethingThatUsesreader(t *testing.T) {
//
//		// make and configure a mocked reader
//		mockedreader := &readerMock{
//			ReadFunc: func(p []byte) (int, error) {
//				panic("mock out the Read method")
//			},
//		}
//
//		// use mockedreader in code that requires reader
//		// and then make assertions.
//
//	}
type readerMock struct {
	// ReadFunc mocks the Read method.
	ReadFunc func(p []byte) (int, error)

	// calls tracks calls to the methods.
	calls struct {
		// Read holds details about calls to the Read method.
		Read []struct {
			// P is the p argument value.
			P []byte
		}
	}
	lockRead sync.RWMutex
}

// Read calls ReadFunc.
func (mock *readerMock) Read(p []byte) (int, error) {
	if mock.ReadFunc == nil {
		panic("readerMock.ReadFunc: method is nil but reader.Read was just called")
	}
	callInfo := struct {
		P []byte
	}{
		P: p,
	}
	mock.lockRead.Lock()
	mock.calls.Read = append(mock.calls.Read, callInfo)
	mock.lockRead.Unlock()
	return mock.ReadFunc(p)
}

// ReadCalls gets all the calls that were made to Read.
// Check the length with:
//
//	len(mockedreader.ReadCalls())
func (mock *readerMock) ReadCalls() []struct {
	P []byte
} {
	var calls []struct {
		P []byte
	}
	mock.lockRead.RLock()
	calls = mock.calls.Read
	mock.lockRead.RUnlock()
	return calls
}
