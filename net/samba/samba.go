package samba

import (
	"github.com/hirochachacha/go-smb2"
	"io"
	"net"
	"os"
	"time"
)

type Samba struct {
	ServerName    string
	UserName      string
	Password      string
	ShareName     string
	Session       *smb2.Session
	Share         *smb2.Share
	NetConnection func(serverName string) (net.Conn, error)
	SambaDialer   func(userName, password string) *smb2.Dialer
	SambaSession  func(conn net.Conn, dialer *smb2.Dialer) (*smb2.Session, error)
}

func DefaultNetConnection(serverName string) (net.Conn, error) {
	return net.DialTimeout("tcp", serverName, 30*time.Second)
}

func DefaultSambaDialer(userName, password string) *smb2.Dialer {
	return &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     userName,
			Password: password,
		},
	}
}

func DefaultSambaSession(conn net.Conn, dialer *smb2.Dialer) (*smb2.Session, error) {
	return dialer.Dial(conn)
}

func (s *Samba) Connect() error {
	conn, err := NewTCPConnection(s.ServerName)
	if err != nil {
		return err
	}
	dialer := NewDialer(s.UserName, s.Password)
	session, err := NewSession(conn, dialer)
	if err != nil {
		return err
	}
	s.Session = session
	defer func(Session *smb2.Session) {
		_ = Session.Logoff()
	}(s.Session)
	return nil
}

func (s *Samba) Mount() error {
	share, err := s.Session.Mount(s.ShareName)
	if err != nil {
		return err
	}
	s.Share = share
	defer func(Share *smb2.Share) {
		_ = Share.Umount()
	}(s.Share)
	return nil
}

func (s *Samba) Send(source string, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func(sourceFile *os.File) {
		_ = sourceFile.Close()
	}(sourceFile)

	destinationFile, err := s.Share.Create(destination)
	if err != nil {
		return err
	}
	defer func(destinationFile *smb2.File) {
		_ = destinationFile.Close()
	}(destinationFile)

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}

func (s *Samba) SendMany(sourcePath string, destinationPath string, fileNames []string) error {
	for _, fileName := range fileNames {
		err := s.Send(sourcePath+fileName, destinationPath+fileName)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewTCPConnection(server string) (*net.Conn, error) {
	conn, err := net.DialTimeout("tcp", server, time.Second*30)
	if err != nil {
		return nil, err
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	return &conn, nil
}

func NewDialer(username string, password string) *smb2.Dialer {
	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
		},
	}
	return dialer
}

func NewSession(conn *net.Conn, dialer *smb2.Dialer) (*smb2.Session, error) {
	session, err := dialer.Dial(*conn)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func NewSamba(serverName, userName, password, shareName string) *Samba {
	return &Samba{
		ServerName:    serverName,
		UserName:      userName,
		Password:      password,
		ShareName:     shareName,
		NetConnection: DefaultNetConnection,
		SambaDialer:   DefaultSambaDialer,
		SambaSession:  DefaultSambaSession,
	}
}
