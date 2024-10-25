package samba

import (
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/hirochachacha/go-smb2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockNetConn is a mock implementation of the net.Conn interface
type MockNetConn struct {
	mock.Mock
	net.Conn
}

// MockSMBSession is a mock implementation of the smb2.Session
type MockSMBSession struct {
	mock.Mock
	smb2.Session
}

// MockSambaDialer creates a new SMB dialer
func MockSambaDialer(userName, password string) *smb2.Dialer {
	return &smb2.Dialer{}
}

// MockNetConnection is a mock implementation of network connection
func MockNetConnection(serverName string) (net.Conn, error) {
	conn := MockNetConn{}
	args := conn.Called(serverName)
	return args.Get(0).(net.Conn), args.Error(1)
}

// MockSambaSession is a mock implementation of Samba session
func MockSambaSession(conn net.Conn, dialer *smb2.Dialer) (*smb2.Session, error) {
	mockSMBSession := MockSMBSession{}
	args := mockSMBSession.Called(conn, dialer)
	return args.Get(0).(*smb2.Session), args.Error(1)
}

func mockServer(t *testing.T, wg *sync.WaitGroup, address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address) // ":0" lets OS pick a free port
	if err != nil {
		return nil, fmt.Errorf("failed to listen on dynamic port: %w", err)
	}
	t.Logf("Mock server started at %s", address)

	go func() {
		defer wg.Done() // Indicate goroutine completion to WaitGroup
		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Logf("Error accepting connection: %v", err)
				return
			}
			// Handle the connection (e.g., read/write data)
			fmt.Fprintf(conn, "Hello\n")
			conn.Close()
		}
	}()

	fmt.Printf("Mock server started at %s\n", listener.Addr().String())

	return listener, nil
}

// TestConnect tests the Connect function of the Samba struct
func TestConnect(t *testing.T) {
	mockConn := new(MockNetConn)
	mockSession := new(MockSMBSession)
	mockDialer := &smb2.Dialer{}

	// Set up the mock functions with testify
	mockConn.On("Dial", "localhost:445").Return(mockConn, nil)
	mockSession.On("Dial", mockConn, mockDialer).Return(mockSession, nil)
	mockSession.On("Logoff").Return(nil)

	// Create a mock Samba instance and inject mock functions
	mockSamba := &Samba{
		Host:          "127.0.0.1",
		Port:          5001,
		UserName:      "test-user",
		Password:      "test-pass",
		ShareName:     "test-share",
		NetConnection: MockNetConnection,
		SambaDialer:   MockSambaDialer,
		SambaSession:  MockSambaSession,
	}

	var wg sync.WaitGroup
	wg.Add(1) // Add to WaitGroup for synchronization

	listener, err := mockServer(t, &wg, mockSamba.Server())
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	// Run the Connect method
	err = mockSamba.Connect()
	require.NoError(t, err)

	// Assert all mock expectations
	mockSession.AssertExpectations(t)
	mockConn.AssertExpectations(t)

	// Indicate to WaitGroup that the work is done
	wg.Wait()
}
