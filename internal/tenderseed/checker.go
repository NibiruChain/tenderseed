package tenderseed

import (
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/pex"
)

const (
	defaultAttemptsLimit = 3
)

type PeerChecker struct {
	addrBook pex.AddrBook
	done     chan bool
	log      log.Logger
	attempts map[string]int
}

func NewPeerChecker(addrBook pex.AddrBook, log log.Logger) *PeerChecker {
	return &PeerChecker{
		addrBook: addrBook,
		log:      log,
		attempts: make(map[string]int),
	}
}

func (s *PeerChecker) Start() {
	s.done = make(chan bool)

	go func() {
		for {
			select {
			case <-s.done:
				return
			default:
				s.processPeers()
			}
		}
	}()
}

func (s *PeerChecker) Stop() {
	s.done <- true
}

func (s *PeerChecker) processPeers() {
	s.log.Info("start evaluating peers")

	for _, peer := range s.addrBook.GetSelection() {
		select {
		case <-s.done:
			return
		default:
			if peer != nil {
				if conn, err := peer.DialTimeout(time.Second); err != nil {
					s.addAttempt(peer)
					s.addrBook.MarkAttempt(peer)
					if s.reachedAttemptsLimit(peer) {
						s.log.Info("marking peer as bad", "peer", peer, "attempts", s.attemptsNumber(peer))
						s.resetAttempts(peer)
						s.addrBook.MarkBad(peer, time.Hour)
					}
				} else {
					s.resetAttempts(peer)
					conn.Close()
					s.log.Info("marking peer as good", "peer", peer, "attempts", s.attemptsNumber(peer))
					s.addrBook.MarkGood(peer.ID)
				}
			}
		}

	}
}

func (s *PeerChecker) addAttempt(peer *p2p.NetAddress) {
	if _, ok := s.attempts[string(peer.ID)]; !ok {
		s.resetAttempts(peer)
	}
	s.attempts[string(peer.ID)] = s.attempts[string(peer.ID)] + 1
}

func (s *PeerChecker) reachedAttemptsLimit(peer *p2p.NetAddress) bool {
	if _, ok := s.attempts[string(peer.ID)]; !ok {
		s.resetAttempts(peer)
		return false
	}
	return s.attempts[string(peer.ID)] >= defaultAttemptsLimit
}

func (s *PeerChecker) resetAttempts(peer *p2p.NetAddress) {
	s.attempts[string(peer.ID)] = 0
}

func (s *PeerChecker) attemptsNumber(peer *p2p.NetAddress) int {
	if _, ok := s.attempts[string(peer.ID)]; !ok {
		return 0
	}
	return s.attempts[string(peer.ID)]
}
