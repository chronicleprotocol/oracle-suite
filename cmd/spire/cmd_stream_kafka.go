//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/cmd"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
)

func NewStreamKafkaCmd(cfg *spire.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	var kafkaTopic string
	cmd := &cobra.Command{
		Use:   "kafka --topic kafka_topic [LIBP2P_TOPIC...]",
		Args:  cobra.MinimumNArgs(0),
		Short: "Streams data from the network into kafka topics",
		RunE: func(cmd *cobra.Command, topics []string) (err error) {
			if err := cf.Load(cfg); err != nil {
				return err
			}
			logger := lf.Logger()
			if len(topics) == 0 {
				topics = messages.AllMessagesMap.Keys()
			}
			if kafkaTopic == "" {
				return fmt.Errorf("kafka topic is required")
			}
			services, err := cfg.StreamServices(logger, cmd.Root().Use, cmd.Root().Version, topics...)
			if err != nil {
				return err
			}
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			if err = services.Start(ctx); err != nil {
				return err
			}
			defer func() {
				ctxCancel()
				if sErr := <-services.Wait(); err == nil {
					err = sErr
				}
			}()
			sink := chanutil.NewFanIn[transport.ReceivedMessage]()
			for _, s := range topics {
				ch := services.Transport.Messages(s)
				if ch == nil {
					return fmt.Errorf("unconfigured topic: %s", s)
				}
				if err := sink.Add(ch); err != nil {
					return err
				}
				logger.
					WithField("name", s).
					Info("Subscribed to topic")
			}

			brokersCfg := strings.Split(cfg.Kafka.Brokers, ",")
			producer, err := startKafkaProducer(brokersCfg)
			if producer == nil {
				return fmt.Errorf("failed to setup producer: %w", err)
			}

			sinkCh := sink.Chan()
			for {
				select {
				case <-ctx.Done():
					return nil
				case msg, ok := <-sinkCh:
					if !ok {
						return nil
					}
					jsonMsg, err := json.Marshal(handleMessage(msg))
					if err != nil {
						lf.Logger().WithError(err).Error("Failed to marshal message")
						continue
					}
					err = sendKafkaMessage(producer, kafkaTopic, jsonMsg)
					if err != nil {
						lf.Logger().WithError(err).Error("Failed to send message to kafka")
						continue
					}
					// fmt.Println(string(jsonMsg))
				}
			}
		},
	}
	cmd.AddCommand(
		NewStreamPricesCmd(cfg, cf, lf),
		NewTopicsCmd(),
	)
	cmd.Flags().StringVar(
		&kafkaTopic,
		"topic",
		"",
		"topic to publish to",
	)
	var format string
	cmd.Flags().StringVarP(&format, "output", "o", "", "(here for backward compatibility)")
	return cmd
}

func startKafkaProducer(brokers []string) (sarama.SyncProducer, error) {
	config := kafkaConfig()
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to setup kafka producer: %w", err)
	}
	return producer, nil
}

func sendKafkaMessage(producer sarama.SyncProducer, topic string, data []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		// Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(data),
	}

	_, _, err := producer.SendMessage(msg)
	return err
}

func kafkaConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.Idempotent = true
	config.Producer.Return.Errors = false
	config.Producer.Return.Successes = true
	// config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Transaction.Retry.Backoff = 10
	// config.Producer.Transaction.ID = "txn_producer"
	config.Net.MaxOpenRequests = 1
	return config
}
