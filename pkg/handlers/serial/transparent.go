package serial

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
)

type TransparentServer struct {
	listener   net.Listener
	serialPort *os.File
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

