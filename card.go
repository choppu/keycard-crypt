package main

import (
	"fmt"

	"github.com/ebfe/scard"
)

func waitForCard(ctx *scard.Context, readers []string) (int, error) {
	rs := make([]scard.ReaderState, len(readers))

	for i := range rs {
		rs[i].Reader = readers[i]
		rs[i].CurrentState = scard.StateUnaware
	}

	for {
		for i := range rs {
			if rs[i].EventState&scard.StatePresent != 0 {
				return i, nil
			}

			rs[i].CurrentState = rs[i].EventState
		}

		err := ctx.GetStatusChange(rs, -1)
		if err != nil {
			return -1, err
		}
	}
}

func createContext() *scard.Context {
	ctx, err := scard.EstablishContext()
	if err != nil {
		fail("error establishing card context")
	}

	return ctx
}

func releaseContext(ctx *scard.Context) {
	if err := ctx.Release(); err != nil {
		fmt.Println("error releasing context")
	}
}

func connectCard(ctx *scard.Context) *scard.Card {
	readers, err := ctx.ListReaders()
	if err != nil {
		fail("error getting readers")
	}

	fmt.Println("waiting for a card")
	if len(readers) == 0 {
		fail("no smartcard reader found")
	}

	index, err := waitForCard(ctx, readers)
	if err != nil {
		fail("error waiting for card")
	}

	fmt.Println("card found: ", index)
	reader := readers[index]

	fmt.Println("using reader: ", reader)
	card, err := ctx.Connect(reader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		fail("error connecting to card")
	}

	return card
}

func disconnectCard(card *scard.Card) {
	if err := card.Disconnect(scard.ResetCard); err != nil {
		fmt.Println("error disconnecting card")
	}
}

func getCardStatus(card *scard.Card) {
	status, err := card.Status()
	if err != nil {
		fail("error getting card status")
	}

	switch status.ActiveProtocol {
	case scard.ProtocolT0:
		fmt.Println("card protocol T0")
	case scard.ProtocolT1:
		fmt.Println("card protocol T1")
	default:
		fmt.Println("card protocol unknown")
	}
}
