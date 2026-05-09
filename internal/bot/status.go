package bot

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"botucp/internal/config"

	"github.com/bwmarrin/discordgo"
)

// sampInfo menyimpan hasil query SA-MP server
type sampInfo struct {
	Online     int
	MaxPlayers int
}

// querySAMP melakukan query UDP ke SA-MP server
func querySAMP(host, port string) (*sampInfo, error) {
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("udp", addr, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(3 * time.Second))

	ip := net.ParseIP(host).To4()
	if ip == nil {
		ips, err := net.LookupHost(host)
		if err != nil || len(ips) == 0 {
			return nil, fmt.Errorf("gagal resolve host")
		}
		ip = net.ParseIP(ips[0]).To4()
	}

	portNum := uint16(7777)
	fmt.Sscanf(port, "%d", &portNum)

	packet := make([]byte, 11)
	copy(packet[0:4], []byte("SAMP"))
	copy(packet[4:8], ip)
	binary.LittleEndian.PutUint16(packet[8:10], portNum)
	packet[10] = 'i'

	if _, err := conn.Write(packet); err != nil {
		return nil, err
	}

	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil || n < 11 {
		return nil, fmt.Errorf("response tidak valid")
	}

	data := buf[11:]
	if len(data) < 5 {
		return nil, fmt.Errorf("data response terlalu pendek")
	}

	online := int(binary.LittleEndian.Uint16(data[1:3]))
	maxPlayers := int(binary.LittleEndian.Uint16(data[3:5]))

	return &sampInfo{
		Online:     online,
		MaxPlayers: maxPlayers,
	}, nil
}

// startStatusUpdater memulai goroutine untuk update status bot setiap 30 detik
func (b *Bot) startStatusUpdater() {
	update := func() {
		sc := config.LoadServerConfig(b.cfg)
		info, err := querySAMP(sc.Host, sc.Port)
		if err != nil {
			b.session.UpdateStatusComplex(discordgo.UpdateStatusData{
				Activities: []*discordgo.Activity{
					{
						Name: sc.Name + " — Offline",
						Type: discordgo.ActivityTypeWatching,
					},
				},
				Status: "idle",
			})
			return
		}
		b.session.UpdateStatusComplex(discordgo.UpdateStatusData{
			Activities: []*discordgo.Activity{
				{
					Name: fmt.Sprintf("%s — %d/%d", sc.Name, info.Online, info.MaxPlayers),
					Type: discordgo.ActivityTypeWatching,
				},
			},
			Status: "online",
		})
	}

	update()
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			update()
		}
	}()
}
