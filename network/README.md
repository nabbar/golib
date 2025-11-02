# Network package
This package is useful to retrieve network interface information and counter.

The information available are : 
- `name` of the interface (cannot be empty)
- `physical` address (or hardware address) (can be empty for example virtual ip or bridge)
- `MTU` (Maximu Transmission Unit) is the maximum size of packets could be transmit
- `Addr` is the list of all network address (can be empty)
- `Index` is the index for the current interface into the list all interface for current os
- `Flags` are capabilities supported by the interface. You can find const for all flags available in golang/x/net package :
  - `FlagUp`:  interface is up
  - `FlagBroadcast`: interface supports broadcast access capability
  - `FlagLoopback`: interface is a loopback interface
  - `FlagPointToPoint`: interface belongs to a point-to-point link
  - `FlagMulticast`: interface supports multicast access capability
 

## Example of implement
You can get all interface with filter. 

In this example we will search all interface who's have physical address and having at least on address =and match the list of flags composed here with only FlagUp :
```go
// GetAllInterfaces(context.Context, onlyPhysical, hasAddr bool, atLeastMTU int, withFlags ...net.Flags)
list, err := network.GetAllInterfaces(context.Background(), true, true, 0, net.FlagUp)
```


You can also retrieve on interface based on his `name` or his `phyisical` address : 
```go
//NewInterface(context.Context, name, physical string)
i, e := NewInterface(context.Context, "eth0", "") 
``` 

When you have you interface variable, you can use it to print for example the counter of received data.

The available statistics are grouped by way : In or Out
In each group, the stats are : 
- `StatBytes` : Traffic in Bytes
- `StatPackets` : Number of Packets
- `StatFifo` : Number of packet in FIFO 
- `StatDrop` : Number of packet dropped
- `StatErr` : Number of packet in Error

Stats are given as `map[Stats]Number` : 
- `Stats` is a self defined to map constant with a statement. 
- `Number` are still self defined type to allow formatting. 
- `Bytes` is closed similar to Number but with different unit 

This translation between bytes and number included into each type : 
- `Number.AsBytes`
- `Bytes.AsNumber`

You can also convert this type as `uint64` or as `float64`.

Now, we have all information to understand this example who's print all stat of sending traffic with unit :     
```go
		fmt.Print("\tOut: \n")
 		l = i.GetStatOut()
 		for _, k := range network.ListStatsSort() {
 			s := network.Stats(k)
 			if v, ok := l[s]; ok {
 				fmt.Printf("\t\t%s\n", s.FormatLabelUnitPadded(v))
 			} else {
 				fmt.Printf("\t\t%s\n", s.FormatLabelUnitPadded(0))
 			}
 		}

```

## Example 
You can find an example in [`test/test-network`](../zzz-test-standalone/test-network/main.go)

There result of this example is like this : 
```shell
iface 'eth0' [00:15:5d:66:60:dd] :
        Flags: up broadcast multicast
        Addrs: {"addr":"172.25.59.95/20"} {"addr":"fe80::215:5dff:fe66:60dd/64"}
        In:
                Traffic:   27.33 GB
                Packets:   20 M
                Fifo:       0
                Drop:       0
                Error:      0
        Out:
                Traffic:   28.03 GB
                Packets: 1118 K
                Fifo:       0
                Drop:       0
                Error:      0
```

