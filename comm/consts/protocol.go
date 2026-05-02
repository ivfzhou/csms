/*
 * Copyright (c) 2024 ivfzhou
 * csms is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package consts

const (
	WHQLConfigOff = 1
	WHQLConfigOn  = 2
)

type WHQLConfig struct {
	ELAMConfig                            `json:"elamConfig"`
	WFPConfig                             `json:"wfpConfig"`
	IsWindowsDriverProject                bool `json:"isWindowsDriverProject,omitempty"`
	IsHSPCompatibility                    bool `json:"isHSPCompatibility"`
	AudioCodecVerifyAudioEffectsDiscovery bool `json:"audioCodecVerifyAudioEffectsDiscovery,omitempty"`
}

type ELAMConfig struct {
	IsWdBootMVIMember int `json:"isWdBootMVIMember,omitempty"`
}

type WFPConfig struct {
	ProductName                       string `json:"productName,omitempty"`
	DriverName                        string `json:"driverName,omitempty"`
	CompanyName                       string `json:"companyName,omitempty"`
	EnableDriverVerifier              int    `json:"enableDriverVerifier,omitempty"`
	CalloutDriver                     int    `json:"calloutDriver,omitempty"`
	IsAFirewall                       int    `json:"isAFirewall,omitempty"`
	LayeredOnMicrosoftWindowsFirewall int    `json:"layeredOnMicrosoftWindowsFirewall,omitempty"`
	DoesMACFiltering                  int    `json:"doesMACFiltering,omitempty"`
	DoesVSwitchFiltering              int    `json:"doesVSwitchFiltering,omitempty"`
	DoesPacketInjection               int    `json:"doesPacketInjection,omitempty"`
	DoesStreamInjection               int    `json:"doesStreamInjection,omitempty"`
	DoesConnectionProxying            int    `json:"doesConnectionProxying,omitempty"`
	SupportModernApplications         int    `json:"supportModernApplications,omitempty"`
	CleanUninstall                    int    `json:"cleanUninstall,omitempty"`
	NoProxyDeadlocks                  int    `json:"noProxyDeadlocks,omitempty"`
	IdentifyingProvider               int    `json:"identifyingProvider,omitempty"`
	AssociateProvider                 int    `json:"associateProvider,omitempty"`
	TerminatingFilter                 int    `json:"terminatingFilter,omitempty"`
	UseOwnSubLayer                    int    `json:"useOwnSubLayer,omitempty"`
	MaintainHelperClass               int    `json:"maintainHelperClass,omitempty"`
	NoAVs                             int    `json:"noAVs,omitempty"`
	NonTampered3rdPartyObjects        int    `json:"nonTampered3rdPartyObjects,omitempty"`
	NoPacketInjectionDeadlocks        int    `json:"noPacketInjectionDeadlocks,omitempty"`
	NoStreamStarvation                int    `json:"noStreamStarvation,omitempty"`
	SupportPowerManagement            int    `json:"supportPowerManagement,omitempty"`
	WFPObjectEnumAndACLs              int    `json:"wfpObjectEnumAndACLs,omitempty"`
	MaxWinSock                        int    `json:"maxWinSock,omitempty"`
	ProperlyDisableWindowsFirewall    int    `json:"properlyDisableWindowsFirewall,omitempty"`
	NoPermitBlockAll                  int    `json:"noPermitBlockAll,omitempty"`
	SupportTupleExceptions            int    `json:"supportTupleExceptions,omitempty"`
	SupportAppExceptions              int    `json:"supportAppExceptions,omitempty"`
	SupportMACAddressExceptions       int    `json:"supportMACAddressExceptions,omitempty"`
	UseWFP                            int    `json:"useWFP,omitempty"`
	SupportARP                        int    `json:"supportARP,omitempty"`
	SupportNeighborDiscovery          int    `json:"supportNeighborDiscovery,omitempty"`
	SupportDHCP                       int    `json:"supportDHCP,omitempty"`
	SupportIPv4                       int    `json:"supportIPv4,omitempty"`
	SupportIPv6                       int    `json:"supportIPv6,omitempty"`
	SupportDNS                        int    `json:"supportDNS,omitempty"`
	Support6To4                       int    `json:"support6To4,omitempty"`
	SupportAutomaticUpdates           int    `json:"supportAutomaticUpdates,omitempty"`
	SupportBasicWebsiteBrowsing       int    `json:"supportBasicWebsiteBrowsing,omitempty"`
	SupportFileAndPrinterSharing      int    `json:"supportFileAndPrinterSharing,omitempty"`
	SupportICMPErrorMessages          int    `json:"supportICMPErrorMessages,omitempty"`
	SupportInternetStreaming          int    `json:"supportInternetStreaming,omitempty"`
	SupportMediaExtenderStreaming     int    `json:"supportMediaExtenderStreaming,omitempty"`
	SupportMobileBroadBand            int    `json:"supportMobileBroadBand,omitempty"`
	SupportPeerNameResolution         int    `json:"supportPeerNameResolution,omitempty"`
	SupportRemoteAssistance           int    `json:"supportRemoteAssistance,omitempty"`
	SupportRemoteDesktop              int    `json:"supportRemoteDesktop,omitempty"`
	SupportTeredo                     int    `json:"supportTeredo,omitempty"`
	SupportVirtualPrivateNetworking   int    `json:"supportVirtualPrivateNetworking,omitempty"`
	InteropWithOtherExtensions        int    `json:"interopWithOtherExtensions,omitempty"`
	NoEgressModification              int    `json:"noEgressModification,omitempty"`
	SupportLiveMigration              int    `json:"supportLiveMigration,omitempty"`
	SupportRemoval                    int    `json:"supportRemoval,omitempty"`
	SupportReordering                 int    `json:"supportReordering,omitempty"`
	InteropWithWFPSampler             int    `json:"interopWithWFPSampler,omitempty"`
}
