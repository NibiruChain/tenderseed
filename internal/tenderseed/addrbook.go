package tenderseed

import (
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/pex"
)

type AddrBook struct {
	pex.AddrBook
}

func NewTenderseedAddrBook(filePath string, routabilityStrict bool) pex.AddrBook {
	am := pex.NewAddrBook(filePath, routabilityStrict)
	return &AddrBook{am}
}

func (a AddrBook) GetSelection() []*p2p.NetAddress {
	selection := make([]*p2p.NetAddress, 0)
	for _, peer := range a.AddrBook.GetSelection() {
		if peer != nil {
			selection = append(selection, peer)
		}
	}
	return selection
}
