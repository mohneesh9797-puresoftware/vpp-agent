//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp1904

import (
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"

	"go.ligato.io/cn-infra/v2/utils/addrs"
	api "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

// AddSPD implements IPSec handler.
func (h *IPSecVppHandler) AddSPD(spdID uint32) error {
	return h.spdAddDel(spdID, true)
}

// DeleteSPD implements IPSec handler.
func (h *IPSecVppHandler) DeleteSPD(spdID uint32) error {
	return h.spdAddDel(spdID, false)
}

// AddSP implements IPSec handler.
func (h *IPSecVppHandler) AddSP(sp *ipsec.SecurityPolicy) error {
	return h.spdAddDelEntry(sp, true)
}

// DeleteSP implements IPSec handler.
func (h *IPSecVppHandler) DeleteSP(sp *ipsec.SecurityPolicy) error {
	return h.spdAddDelEntry(sp, false)
}

// AddSPDInterface implements IPSec handler.
func (h *IPSecVppHandler) AddSPDInterface(spdID uint32, ifaceCfg *ipsec.SecurityPolicyDatabase_Interface) error {
	ifaceMeta, found := h.ifIndexes.LookupByName(ifaceCfg.Name)
	if !found {
		return errors.New("failed to get interface metadata")
	}
	return h.interfaceAddDelSpd(spdID, ifaceMeta.SwIfIndex, true)
}

// DeleteSPDInterface implements IPSec handler.
func (h *IPSecVppHandler) DeleteSPDInterface(spdID uint32, ifaceCfg *ipsec.SecurityPolicyDatabase_Interface) error {
	ifaceMeta, found := h.ifIndexes.LookupByName(ifaceCfg.Name)
	if !found {
		return errors.New("failed to get interface metadata")
	}
	return h.interfaceAddDelSpd(spdID, ifaceMeta.SwIfIndex, false)
}

// AddSA implements IPSec handler.
func (h *IPSecVppHandler) AddSA(sa *ipsec.SecurityAssociation) error {
	return h.sadAddDelEntry(sa, true)
}

// DeleteSA implements IPSec handler.
func (h *IPSecVppHandler) DeleteSA(sa *ipsec.SecurityAssociation) error {
	return h.sadAddDelEntry(sa, false)
}

// AddTunnelProtection implements IPSec handler for adding a tunnel protection.
func (h *IPSecVppHandler) AddTunnelProtection(tp *ipsec.TunnelProtection) error {
	return fmt.Errorf("%w in VPP 19.04", vppcalls.ErrTunnelProtectionUnsupported)
}

// UpdateTunnelProtection implements IPSec handler for updating a tunnel protection.
func (h *IPSecVppHandler) UpdateTunnelProtection(tp *ipsec.TunnelProtection) error {
	return fmt.Errorf("%w in VPP 19.04", vppcalls.ErrTunnelProtectionUnsupported)
}

// DeleteTunnelProtection implements IPSec handler for deleting a tunnel protection.
func (h *IPSecVppHandler) DeleteTunnelProtection(tp *ipsec.TunnelProtection) error {
	return fmt.Errorf("%w in VPP 19.04", vppcalls.ErrTunnelProtectionUnsupported)
}

func (h *IPSecVppHandler) spdAddDel(spdID uint32, isAdd bool) error {
	req := &api.IpsecSpdAddDel{
		IsAdd: boolToUint(isAdd),
		SpdID: spdID,
	}
	reply := &api.IpsecSpdAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *IPSecVppHandler) spdAddDelEntry(sp *ipsec.SecurityPolicy, isAdd bool) error {
	req := &api.IpsecSpdEntryAddDel{
		IsAdd: boolToUint(isAdd),
		Entry: api.IpsecSpdEntry{
			SpdID:           sp.SpdIndex,
			Priority:        sp.Priority,
			IsOutbound:      boolToUint(sp.IsOutbound),
			Protocol:        uint8(sp.Protocol),
			RemotePortStart: uint16(sp.RemotePortStart),
			RemotePortStop:  uint16(sp.RemotePortStop),
			LocalPortStart:  uint16(sp.LocalPortStart),
			LocalPortStop:   uint16(sp.LocalPortStop),
			Policy:          api.IpsecSpdAction(sp.Action),
			SaID:            sp.SaIndex,
		},
	}
	if req.Entry.RemotePortStop == 0 {
		req.Entry.RemotePortStop = ^req.Entry.RemotePortStop
	}
	if req.Entry.LocalPortStop == 0 {
		req.Entry.LocalPortStop = ^req.Entry.LocalPortStop
	}

	var err error
	req.Entry.RemoteAddressStart, err = IPToAddress(ipOr(sp.RemoteAddrStart, "0.0.0.0"))
	if err != nil {
		return err
	}
	req.Entry.RemoteAddressStop, err = IPToAddress(ipOr(sp.RemoteAddrStop, "255.255.255.255"))
	if err != nil {
		return err
	}
	req.Entry.LocalAddressStart, err = IPToAddress(ipOr(sp.LocalAddrStart, "0.0.0.0"))
	if err != nil {
		return err
	}
	req.Entry.LocalAddressStop, err = IPToAddress(ipOr(sp.LocalAddrStop, "255.255.255.255"))
	if err != nil {
		return err
	}

	reply := &api.IpsecSpdEntryAddDelReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}


func ipOr(s, o string) string {
	if s != "" {
		return s
	}
	return o
}

func (h *IPSecVppHandler) interfaceAddDelSpd(spdID, swIfIdx uint32, isAdd bool) error {
	req := &api.IpsecInterfaceAddDelSpd{
		IsAdd:     boolToUint(isAdd),
		SwIfIndex: swIfIdx,
		SpdID:     spdID,
	}
	reply := &api.IpsecInterfaceAddDelSpdReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func (h *IPSecVppHandler) sadAddDelEntry(sa *ipsec.SecurityAssociation, isAdd bool) error {
	cryptoKey, err := hex.DecodeString(sa.CryptoKey)
	if err != nil {
		return err
	}
	integKey, err := hex.DecodeString(sa.IntegKey)
	if err != nil {
		return err
	}

	var flags api.IpsecSadFlags
	if sa.UseEsn {
		flags |= api.IPSEC_API_SAD_FLAG_USE_ESN
	}
	if sa.UseAntiReplay {
		flags |= api.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY
	}
	if sa.EnableUdpEncap {
		flags |= api.IPSEC_API_SAD_FLAG_UDP_ENCAP
	}
	var tunnelSrc, tunnelDst api.Address
	if sa.TunnelSrcAddr != "" {
		flags |= api.IPSEC_API_SAD_FLAG_IS_TUNNEL
		isIPv6, err := addrs.IsIPv6(sa.TunnelSrcAddr)
		if err != nil {
			return err
		}
		if isIPv6 {
			flags |= api.IPSEC_API_SAD_FLAG_IS_TUNNEL_V6
		}
		tunnelSrc, err = IPToAddress(sa.TunnelSrcAddr)
		if err != nil {
			return err
		}
		tunnelDst, err = IPToAddress(sa.TunnelDstAddr)
		if err != nil {
			return err
		}
	}

	req := &api.IpsecSadEntryAddDel{
		IsAdd: boolToUint(isAdd),
		Entry: api.IpsecSadEntry{
			SadID:           sa.Index,
			Spi:             sa.Spi,
			Protocol:        api.IpsecProto(sa.Protocol),
			CryptoAlgorithm: api.IpsecCryptoAlg(sa.CryptoAlg),
			CryptoKey: api.Key{
				Data:   cryptoKey,
				Length: uint8(len(cryptoKey)),
			},
			IntegrityAlgorithm: api.IpsecIntegAlg(sa.IntegAlg),
			IntegrityKey: api.Key{
				Data:   integKey,
				Length: uint8(len(integKey)),
			},
			TunnelSrc: tunnelSrc,
			TunnelDst: tunnelDst,
			Flags:     flags,
		},
	}
	reply := &api.IpsecSadEntryAddDelReply{}

	if err = h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

func boolToUint(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}
