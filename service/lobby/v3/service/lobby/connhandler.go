package lobby

import (
	"bytes"
	"io"
	"net"
	"strings"

	"wz2100.net/microlobby/shared/component"
)

type gameStruct struct {
	Version        uint32
	Name           []byte `struc:"[64]int32"` // 64 Chars
	DwSize         int32
	DwFlags        int32
	Host           []byte `struc:"[40]int32"` // 40 chars
	MaxPlayers     int32
	CurrentPlayers int32
	DwUserFlag0    int32
	DwUserFlag1    int32
	DwUserFlag2    int32
	DwUserFlag3    int32
	Host2          []byte `struc:"[40]int32"`  // 40 chars
	Host3          []byte `struc:"[40]int32"`  // 40 chars
	Extra          []byte `struc:"[157]int32"` // 157 chars
	Port           uint16
	MapName        []byte `struc:"[40]int32"`  // 40 chars
	HostName       []byte `struc:"[40]int32"`  // 40 chars
	VersionString  []byte `struc:"[64]int32"`  // 64 chars
	ModList        []byte `struc:"[255]int32"` // 255 chars
	VersionMajor   uint32
	VersionMinor   uint32
	Private        uint32
	Pure           uint32
	Mods           uint32
	GameId         uint32
	Future2        uint32
	Future3        uint32
	Future4        uint32
}

const gameStructSize = 4 + 64 + (4 * 2) + 40 + (4 * 6) + 40 + 40 + 157 + 2 + 40 + 40 + 64 + 255 + (4 * 9)

func stringToPaddedByte(in string, size int) []byte {
	out := make([]byte, size)
	if len(in) > size {
		copy(out, []byte(in[0:size-1]))
	} else {
		copy(out, []byte(in))
	}

	return out
}

func byteToString(in []byte) string {
	return string(bytes.Trim(in, "\x00"))
}

type ConnHandler struct {
	cRegistry *component.Registry
	logrus    component.LogrusComponent
	conn      net.Conn
	closing   bool
}

func NewConnHandler(cRegistry *component.Registry, conn net.Conn) (*ConnHandler, error) {
	logrus, err := component.Logrus(cRegistry)
	if err != nil {
		return nil, err
	}

	return &ConnHandler{cRegistry: cRegistry, logrus: logrus, conn: conn, closing: false}, nil
}

func (h *ConnHandler) Serve() {
	myLogger := h.logrus.WithClassFunc(pkgPath, "ConnHandler", "Serve").WithField("remote", h.conn.RemoteAddr().String())
	myLogger.Info("Got a connection")
	for !h.closing {
		cmdBuf := make([]byte, 5)
		n, err := h.conn.Read(cmdBuf)
		if err != nil {
			// Do not log EOF
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				h.closing = true
				break
			}

			myLogger.Error(err)
			h.closing = true
			break
		}
		if n != 5 {
			myLogger.Errorf("Can't read 5 bytes command String? Got '%v'(%d)", cmdBuf, n)
			h.closing = true
			break
		}

		cmd := strings.ToLower(byteToString(cmdBuf))
		switch cmd {
		case "list":
			myLogger.WithField("cmd", cmd).Trace("Executing")
			break
		case "gaid":
			myLogger.WithField("cmd", cmd).Trace("Executing")
			break
		case "addg":
			myLogger.WithField("cmd", cmd).Trace("Executing")
			break
		default:
			myLogger.WithField("cmd", cmd).Error("Unknown command")
			h.closing = true
			break
		}
	}

	h.conn.Close()
}
