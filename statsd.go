package statsd

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

type StatsdClient struct {
	Addr *net.UDPAddr
}

type Statsd struct {
	Client *StatsdClient
	Prefix string
}

func New(addr string, prefix string) *Statsd {
	udpAddr, _ := net.ResolveUDPAddr("udp", addr)
	client := &StatsdClient{Addr: udpAddr}

	return &Statsd{
		Client: client,
		Prefix: prefix,
	}
}

func (client *StatsdClient) Connection() (conn net.Conn, err error) {
	return net.DialUDP("udp", nil, client.Addr)
}

func (client *StatsdClient) withConnection(fn func(conn *net.Conn)) {
	conn, err := client.Connection()
	if err != nil {
		log.Println(err)
	}

	defer conn.Close()

	fn(&conn)
}

func (s *Statsd) prefixStat(stat string) string {
	return fmt.Sprintf("%s.%s", s.Prefix, stat)
}

func (s *Statsd) Timing(stat string, time int64) {
	s.Client.withConnection(func(conn *net.Conn) {
		s.TimingRaw(conn, stat, time)
	})
}

func (s *Statsd) TimingRaw(conn *net.Conn, stat string, time int64) {
	updateString := fmt.Sprintf("%d|ms", time)
	stats := map[string]string{stat: updateString}
	s.Send(conn, stats, 1)
}

func (s *Statsd) TimingWithSampleRate(stat string, time int64, sampleRate float32) {
	s.Client.withConnection(func(conn *net.Conn) {
		s.TimingWithSampleRateRaw(conn, stat, time, sampleRate)
	})
}

func (s *Statsd) TimingWithSampleRateRaw(conn *net.Conn, stat string, time int64, sampleRate float32) {
	updateString := fmt.Sprintf("%d|ms", time)
	stats := map[string]string{stat: updateString}
	s.Send(conn, stats, sampleRate)
}

func (s *Statsd) Increment(stat string) {
	s.Client.withConnection(func(conn *net.Conn) {
		s.IncrementRaw(conn, stat)
	})
}

func (s *Statsd) IncrementRaw(conn *net.Conn, stat string) {
	stats := []string{stat}
	s.UpdateStats(conn, stats, 1, 1)
}

func (s *Statsd) IncrementWithSampling(stat string, sampleRate float32) {
	s.Client.withConnection(func(conn *net.Conn) {
		s.IncrementWithSamplingRaw(conn, stat, sampleRate)
	})
}

func (s *Statsd) IncrementWithSamplingRaw(conn *net.Conn, stat string, sampleRate float32) {
	stats := []string{stat}
	s.UpdateStats(conn, stats[:], 1, sampleRate)
}

func (s *Statsd) Decrement(stat string) {
	s.Client.withConnection(func(conn *net.Conn) {
		s.DecrementRaw(conn, stat)
	})
}

func (s *Statsd) DecrementRaw(conn *net.Conn, stat string) {
	stats := []string{stat}
	s.UpdateStats(conn, stats[:], -1, 1)
}

func (s *Statsd) DecrementWithSampling(stat string, sampleRate float32) {
	s.Client.withConnection(func(conn *net.Conn) {
		s.DecrementWithSamplingRaw(conn, stat, sampleRate)
	})
}

func (s *Statsd) DecrementWithSamplingRaw(conn *net.Conn, stat string, sampleRate float32) {
	stats := []string{stat}
	s.UpdateStats(conn, stats[:], -1, sampleRate)
}

func (s *Statsd) UpdateStats(conn *net.Conn, stats []string, delta int, sampleRate float32) {
	statsToSend := make(map[string]string)
	for _, stat := range stats {
		updateString := fmt.Sprintf("%d|c", delta)
		statsToSend[stat] = updateString
	}
	s.Send(conn, statsToSend, sampleRate)
}

func (s *Statsd) Send(conn *net.Conn, data map[string]string, sampleRate float32) {
	sampledData := make(map[string]string)
	if sampleRate < 1 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		rNum := r.Float32()
		if rNum <= sampleRate {
			for stat, value := range data {
				sampledUpdateString := fmt.Sprintf("%s|@%f", value, sampleRate)
				sampledData[stat] = sampledUpdateString
			}
		}
	} else {
		sampledData = data
	}

	for k, v := range sampledData {
		update_string := fmt.Sprintf("%s:%s", s.prefixStat(k), v)
		_, err := fmt.Fprintf(*conn, update_string)
		if err != nil {
			log.Println(err)
		}
	}
}
