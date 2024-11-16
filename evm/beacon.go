package evm

import (
	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/r3labs/sse"
)

type BeaconEvent struct {
	Slot            uint64 `json:"slot,string"`
	Epoch           uint64
	EpochTransition bool `json:"epoch_transition"`
}

func (c *Client) SubscribeSlots(ctx context.Context, slots chan<- *BeaconEvent) {
	u := c.url + "/eth/v1/events?topics=head"
	c.log.Println("subscribing to", u)
	client := sse.NewClient(u)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := client.SubscribeRaw(func(msg *sse.Event) {
				var ev BeaconEvent
				err := json.Unmarshal(msg.Data, &ev)
				if err != nil {
					log.WithError(err).Error("error unmarshalling beacon event")
					return
				}
				ev.Epoch = ev.Slot / 32
				slots <- &ev
			})
			if err != nil {
				log.WithError(err).Error("error subscribing to beacon event")
			}
		}
	}

}
