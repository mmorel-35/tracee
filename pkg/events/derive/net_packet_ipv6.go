package derive

import (
	"github.com/aquasecurity/tracee/pkg/events"
	"github.com/aquasecurity/tracee/types/trace"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func NetPacketIPv6() deriveFunction {
	return deriveSingleEvent(events.NetPacketIPv6, deriveNetPacketIPv6Args())
}

func deriveNetPacketIPv6Args() deriveArgsFunction {
	return func(event trace.Event) ([]interface{}, error) {
		var ok bool
		var payload []byte

		// sanity checks

		payloadArg := events.GetArg(&event, "payload")
		if payloadArg == nil {
			return nil, noPayloadError()
		}
		if payload, ok = payloadArg.Value.([]byte); !ok {
			return nil, nonByteArgError()
		}
		payloadSize := len(payload)
		if payloadSize < 1 {
			return nil, emptyPayloadError()
		}

		// initial header type

		if event.ReturnValue != AF_INET6 {
			return nil, nil
		}

		// parse packet

		packet := gopacket.NewPacket(
			payload[4:payloadSize], // base event argument is: |sizeof|[]byte|
			layers.LayerTypeIPv6,
			gopacket.Default,
		)
		if packet == nil {
			return []interface{}{}, parsePacketError()
		}

		layer3 := packet.NetworkLayer()

		switch l3 := layer3.(type) {
		case (*layers.IPv6):
			var ipv6 trace.ProtoIPv6
			copyIPv6ToProtoIPv6(l3, &ipv6)

			return []interface{}{
				l3.SrcIP,
				l3.DstIP,
				ipv6,
			}, nil
		}

		return nil, notProtoPacketError("IPv6")
	}
}

//
// IPv6 protocol type conversion (from gopacket layer to trace type)
//

func copyIPv6ToProtoIPv6(l3 *layers.IPv6, proto *trace.ProtoIPv6) {
	proto.Version = l3.Version
	proto.TrafficClass = l3.TrafficClass
	proto.FlowLabel = l3.FlowLabel
	proto.Length = l3.Length
	proto.NextHeader = l3.NextHeader.String()
	proto.HopLimit = l3.HopLimit
	proto.SrcIP = l3.SrcIP.String()
	proto.DstIP = l3.DstIP.String()
}
