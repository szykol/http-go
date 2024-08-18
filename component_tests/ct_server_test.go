package componenttests

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	mock_net "github.com/szykol/http/mocks"
	"github.com/szykol/http/pkg/http"
	"github.com/szykol/http/pkg/log"
	"go.uber.org/mock/gomock"
)

type testF struct {
	sut  *http.Server
	ctx  context.Context
	ctrl *gomock.Controller

	clientConn net.Conn
	serverConn net.Conn

	listenerMock *mock_net.MockListener
}

func setupTest(t *testing.T) testF {
	server := http.NewServer()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)

	logger := log.NewLogger(log.NewZapCfg())
	ctx = log.WithContext(ctx, logger)
	ctrl := gomock.NewController(t)

	t.Cleanup(func() {
		cancel()
		ctrl.Finish()
	})

	serverConn, clientConn := net.Pipe()

	listener := mock_net.NewMockListener(ctrl)
	addr := &net.TCPAddr{}
	listener.EXPECT().Addr().Return(addr)

	return testF{
		sut:          server,
		ctx:          ctx,
		ctrl:         ctrl,
		clientConn:   clientConn,
		serverConn:   serverConn,
		listenerMock: listener,
	}
}

func getData(conn net.Conn) ([]byte, error) {
	var data []byte
	if err := conn.SetReadDeadline(time.Now().Add(time.Microsecond * 100)); err != nil {
		return data, err
	}

	data, err := io.ReadAll(conn)
	if err != nil {
		return data, err
	}

	return data, nil
}

func TestServer(t *testing.T) {
	f := setupTest(t)

	f.listenerMock.EXPECT().Accept().Return(f.serverConn, nil)

	// wait indefinetly to make sure listener does not spin cpu
	f.listenerMock.EXPECT().Accept().DoAndReturn(func() (net.Conn, error) {
		<-f.ctx.Done()
		return nil, fmt.Errorf("some error")
	})

	requestInput := "POST / HTTP/1.1\r\n Host: localhost:4221\r\n User-Agent: curl/8.4.0\r\n Accept: */*\r\nContent-Length: 16\r\n Content-Type: application/json\r\n\r\n{\"test\":\"value\"}"

	go f.sut.Run(f.ctx, f.listenerMock)

	_, _ = f.clientConn.Write([]byte(requestInput))

	data, err := getData(f.clientConn)
	assert.Nil(t, err)
	assert.Equal(t, []byte("HTTP/1.1 404 Not Found\r\n"), data)
}
