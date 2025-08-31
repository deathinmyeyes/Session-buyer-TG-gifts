package giftTypes

import "github.com/gotd/td/tg"

type GiftResult struct {
	GiftID  int64
	Success bool
	Err     error
}

type GiftSummary struct {
	GiftID    int64
	Requested int64
	Success   int64
}

type GiftRequire struct {
	Gift *tg.StarGift
	// Receiver     []string
	ReceiverType []int
	CountForBuy  int64
	Hide         bool
}
