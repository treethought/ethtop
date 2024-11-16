package evm

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	url   string
	wsurl string
	wsc   *ethclient.Client
	log   *log.Entry
}

func NewClient(url, wsurl string) (*Client, error) {
	wsc, err := ethclient.Dial(wsurl)
	if err != nil {
		return nil, err
	}
	l := log.WithField("source", "client")
	return &Client{url: url, wsc: wsc, log: l}, nil
}

func (c *Client) Close() {
	c.wsc.Close()
}

func (c *Client) SubscribeHeads(ctx context.Context, ch chan<- *types.Header) error {
	// c.log.Info("subscribing to new heads")
	sub, err := c.wsc.SubscribeNewHead(ctx, ch)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			return ctx.Err()
		case <-sub.Err():
			log.WithError(err).Error("error, resubscribing")
			sub, err = c.wsc.SubscribeNewHead(ctx, ch)
			if err != nil {
				log.WithError(err).Error("error resubscribing")
				return err
			}
		}
	}
}
