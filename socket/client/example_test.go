/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package client_test

import (
	"fmt"

	libptc "github.com/nabbar/golib/network/protocol"
	sckclt "github.com/nabbar/golib/socket/client"
	sckcfg "github.com/nabbar/golib/socket/config"
)

// Example demonstrates creating a basic TCP client using the factory.
// This is the simplest way to create a socket client.
func Example() {
	// Create client configuration
	cfg := sckcfg.Client{
		Network: libptc.NetworkTCP,
		Address: "localhost:8080",
	}

	// Create client using factory
	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer cli.Close()

	fmt.Println("TCP client created successfully")
	// Output: TCP client created successfully
}

// ExampleNew demonstrates creating a TCP client using the factory.
func ExampleNew() {
	// Create configuration
	cfg := sckcfg.Client{
		Network: libptc.NetworkTCP,
		Address: "localhost:9000",
	}

	// Create client using factory
	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer cli.Close()

	fmt.Printf("Client created successfully\n")
	// Output: Client created successfully
}

// ExampleNew_tcp demonstrates creating a TCP client.
func ExampleNew_tcp() {
	cfg := sckcfg.Client{
		Network: libptc.NetworkTCP,
		Address: "localhost:9001",
	}

	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer cli.Close()

	fmt.Println("TCP client created")
	// Output: TCP client created
}

// ExampleNew_tcp4 demonstrates creating a TCP4 client (IPv4 only).
func ExampleNew_tcp4() {
	cfg := sckcfg.Client{
		Network: libptc.NetworkTCP4,
		Address: "127.0.0.1:9002",
	}

	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer cli.Close()

	fmt.Println("TCP4 client created")
	// Output: TCP4 client created
}

// ExampleNew_tcp6 demonstrates creating a TCP6 client (IPv6 only).
func ExampleNew_tcp6() {
	cfg := sckcfg.Client{
		Network: libptc.NetworkTCP6,
		Address: "[::1]:9003",
	}

	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer cli.Close()

	fmt.Println("TCP6 client created")
	// Output: TCP6 client created
}

// ExampleNew_udp demonstrates creating a UDP client.
func ExampleNew_udp() {
	cfg := sckcfg.Client{
		Network: libptc.NetworkUDP,
		Address: "localhost:9004",
	}

	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer cli.Close()

	fmt.Println("UDP client created")
	// Output: UDP client created
}

// ExampleNew_udp4 demonstrates creating a UDP4 client (IPv4 only).
func ExampleNew_udp4() {
	cfg := sckcfg.Client{
		Network: libptc.NetworkUDP4,
		Address: "127.0.0.1:9005",
	}

	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer cli.Close()

	fmt.Println("UDP4 client created")
	// Output: UDP4 client created
}

// ExampleNew_udp6 demonstrates creating a UDP6 client (IPv6 only).
func ExampleNew_udp6() {
	cfg := sckcfg.Client{
		Network: libptc.NetworkUDP6,
		Address: "[::1]:9006",
	}

	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer cli.Close()

	fmt.Println("UDP6 client created")
	// Output: UDP6 client created
}

// ExampleNew_invalidProtocol demonstrates error handling for invalid protocols.
func ExampleNew_invalidProtocol() {
	// Use an invalid protocol value
	cfg := sckcfg.Client{
		Network: 255, // Invalid protocol
		Address: "localhost:9007",
	}

	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		fmt.Println("Error: invalid protocol")
	}
	_ = cli
	// Output: Error: invalid protocol
}

// ExampleNew_tcpWithTLS demonstrates creating a TCP client with TLS configuration.
func ExampleNew_tcpWithTLS() {
	cfg := sckcfg.Client{
		Network: libptc.NetworkTCP,
		Address: "localhost:9443",
	}
	cfg.TLS.Enabled = true
	cfg.TLS.ServerName = "localhost"

	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer cli.Close()

	fmt.Println("TCP client with TLS created")
	// Output: TCP client with TLS created
}

// Example_multipleClients demonstrates creating multiple clients with different protocols.
func Example_multipleClients() {
	// Create TCP client
	tcpCfg := sckcfg.Client{
		Network: libptc.NetworkTCP,
		Address: "localhost:9100",
	}
	tcpCli, err := sckclt.New(tcpCfg, nil)
	if err != nil {
		fmt.Printf("TCP Error: %v\n", err)
		return
	}
	defer tcpCli.Close()

	// Create UDP client
	udpCfg := sckcfg.Client{
		Network: libptc.NetworkUDP,
		Address: "localhost:9101",
	}
	udpCli, err := sckclt.New(udpCfg, nil)
	if err != nil {
		fmt.Printf("UDP Error: %v\n", err)
		return
	}
	defer udpCli.Close()

	fmt.Println("Multiple clients created successfully")
	// Output: Multiple clients created successfully
}

// Example_errorHandling demonstrates proper error handling patterns.
func Example_errorHandling() {
	cfg := sckcfg.Client{
		Network: libptc.NetworkTCP,
		Address: "localhost:9200",
	}

	cli, err := sckclt.New(cfg, nil)
	if err != nil {
		if err == sckcfg.ErrInvalidProtocol {
			fmt.Println("Protocol not supported")
		} else {
			fmt.Printf("Failed to create client: %v\n", err)
		}
		return
	}
	defer cli.Close()

	fmt.Println("Client created with proper error handling")
	// Output: Client created with proper error handling
}
