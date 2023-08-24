// Package datapoint and its subpackages offer utilities for retrieving and
// manipulating data points.
//
// In this package, a data point represents a value from a specific source, also
// known as 'origin'. This value could represent anything, such as the price of
// an asset at a specific time.
//
// Different data point types are represented as unique types that implement
// the Value interface from the value package. All of these types must support
// binary marshalling, which is essential for transmission across the transport
// layer (see value/types.go).
//
// A provider is a service that produces data points. All providers must
// implement the Provider interface. Currently, the only implementation of
// the Provider interface comes from the graph package.
//
// The graph package uses a graph structure to calculate data points. Instead
// of directly fetching data points from origins, it allows for manipulation,
// aggregation, filtering of data points before delivering them. For instance,
// a graph can calculate the median value from multiple origins, compute
// indirect prices if direct prices are unavailable. For instance, if the
// direct ETH/BTC price is not available, it can be calculated using ETH/USD
// and USD/BTC prices.
//
// A specific graph within this system is often referred to as a 'model'.
// However, within this package, the Model is a simple structure that is
// designed to help users understand how data points are calculated and
// retrieved.
//
// The signer package is responsible for data point signing. A data point may
// be signed by a feed prior to being sent to the transport layer.
//
// The store package facilitates storing of data points received from the
// transport layer.
package datapoint
