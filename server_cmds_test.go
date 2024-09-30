package ts3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCmdsServer(t *testing.T) {
	s := newServer(t)
	defer func() {
		assert.NoError(t, s.Close())
	}()

	c, err := NewClient(s.Addr, Timeout(time.Second*2))
	if !assert.NoError(t, err) {
		return
	}

	defer func() {
		assert.NoError(t, c.Close())
	}()

	testCmdsServer(t, c)
}

func TestCmdsServerSSH(t *testing.T) {
	s := newServer(t, useSSH())
	defer func() {
		assert.NoError(t, s.Close())
	}()

	c, err := NewClient(s.Addr, Timeout(time.Second*2), SSH(sshClientTestConfig))
	if !assert.NoError(t, err) {
		return
	}

	defer func() {
		assert.NoError(t, c.Close())
	}()

	testCmdsServer(t, c)
}

func testCmdsServer(t *testing.T, c *Client) {
	t.Helper()
	list := func(t *testing.T) {
		t.Helper()
		servers, err := c.Server.List()
		if !assert.NoError(t, err) {
			return
		}
		expected := []*Server{
			{
				ID:                 1,
				Port:               10677,
				Status:             "online",
				ClientsOnline:      1,
				QueryClientsOnline: 1,
				MaxClients:         35,
				Uptime:             12345025,
				Name:               "Server #1",
				AutoStart:          true,
				MachineID:          "1",
				UniqueIdentifier:   "uniq1",
			},
			{
				ID:                 2,
				Port:               10617,
				Status:             "online",
				ClientsOnline:      3,
				QueryClientsOnline: 2,
				MaxClients:         10,
				Uptime:             3165117,
				Name:               "Server #2",
				AutoStart:          true,
				MachineID:          "1",
				UniqueIdentifier:   "uniq2",
			},
		}
		assert.Equal(t, expected, servers)
	}

	idgetbyport := func(t *testing.T) {
		t.Helper()
		id, err := c.Server.IDGetByPort(1234)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, 1, id)
	}

	info := func(t *testing.T) {
		t.Helper()
		s, err := c.Server.Info()
		if !assert.NoError(t, err) {
			return
		}
		expected := &Server{
			Status:                                 "template",
			MaxClients:                             32,
			Name:                                   "Test Server",
			AntiFloodPointsNeededCommandBlock:      150,
			AntiFloodPointsNeededIPBlock:           250,
			AntiFloodPointsTickReduce:              5,
			ComplainAutoBanCount:                   5,
			ComplainAutoBanTime:                    1200,
			ComplainRemoveTime:                     3600,
			DefaultChannelAdminGroup:               1,
			DefaultChannelGroup:                    4,
			DefaultServerGroup:                     5,
			MinClientsInChannelBeforeForcedSilence: 100,
			NeededIdentitySecurityLevel:            8,
			LogPermissions:                         true,
			PrioritySpeakerDimmModificator:         -18,
			MaxDownloadTotalBandwidth:              18446744073709551615,
			MaxUploadTotalBandwidth:                18446744073709551615,
			FileBase:                               "files",
			HostButtonToolTip:                      "Multiplay Game Servers",
			HostButtonURL:                          "http://www.multiplaygameservers.com",
			WelcomeMessage:                         "Welcome to TeamSpeak, check [URL]www.teamspeak.com[/URL] for latest infos.",
			VirtualServerDownloadQuota:             18446744073709551615,
			VirtualServerUploadQuota:               18446744073709551615,
		}
		assert.Equal(t, expected, s)
	}

	create := func(t *testing.T) {
		t.Helper()
		s, err := c.Server.Create("my server")
		if !assert.NoError(t, err) {
			return
		}
		expected := &CreatedServer{
			ID:    2,
			Port:  9988,
			Token: "eKnFZQ9EK7G7MhtuQB6+N2B1PNZZ6OZL3ycDp2OW",
		}
		assert.Equal(t, expected, s)
	}

	edit := func(t *testing.T) {
		t.Helper()
		assert.NoError(t, c.Server.Edit(NewArg("virtualserver_maxclients", 10)))
	}

	del := func(t *testing.T) {
		t.Helper()
		assert.NoError(t, c.Server.Delete(1))
	}

	start := func(t *testing.T) {
		t.Helper()
		assert.NoError(t, c.Server.Start(1))
	}

	stop := func(t *testing.T) {
		t.Helper()
		assert.NoError(t, c.Server.Stop(1))
	}

	grouplist := func(t *testing.T) {
		t.Helper()
		groups, err := c.Server.GroupList()
		if !assert.NoError(t, err) {
			return
		}

		expected := []*Group{
			{
				ID:   1,
				Name: "Guest Server Query",
				Type: 2,
			},
			{
				ID:                2,
				Name:              "Admin Server Query",
				Type:              2,
				IconID:            500,
				Saved:             true,
				ModifyPower:       100,
				MemberAddPower:    100,
				MemberRemovePower: 100,
			},
		}
		assert.Equal(t, expected, groups)
	}

	privilegekeylist := func(t *testing.T) {
		t.Helper()
		keys, err := c.Server.PrivilegeKeyList()
		if !assert.NoError(t, err) {
			return
		}

		expected := []*PrivilegeKey{
			{
				Token:   "zTfamFVhiMEzhTl49KrOVYaMilHPgQEBQOJFh6qX",
				ID1:     17395,
				Created: 1499948005,
			},
		}
		assert.Equal(t, expected, keys)
	}

	privilegekeyadd := func(t *testing.T) {
		t.Helper()
		token, err := c.Server.PrivilegeKeyAdd(0, 17395, 0)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "zTfamFVhiMEzhTl49KrOVYaMilHPgQEBQOJFh6qX", token)
	}

	serverrequestconnectioninfo := func(t *testing.T) {
		t.Helper()
		ci, err := c.Server.ServerConnectionInfo()
		if !assert.NoError(t, err) {
			return
		}
		expected := &ServerConnectionInfo{
			FileTransferBandwidthSent:     0,
			FileTransferBandwidthReceived: 0,
			FileTransferTotalSent:         617,
			FileTransferTotalReceived:     0,
			PacketsSentTotal:              926413,
			PacketsReceivedTotal:          650335,
			BytesSentTotal:                92911395,
			BytesReceivedTotal:            61940731,
			BandwidthSentLastSecond:       0,
			BandwidthReceivedLastSecond:   0,
			BandwidthSentLastMinute:       0,
			BandwidthReceivedLastMinute:   0,
			ConnectedTime:                 49408,
			PacketLossTotalAvg:            0.0,
			PingTotalAvg:                  0.0,
			PacketsSentSpeech:             320432180,
			PacketsReceivedSpeech:         174885295,
			BytesSentSpeech:               43805818511,
			BytesReceivedSpeech:           24127808273,
			PacketsSentKeepalive:          55230363,
			PacketsReceivedKeepalive:      55149547,
			BytesSentKeepalive:            2264444883,
			BytesReceivedKeepalive:        2316390993,
			PacketsSentControl:            2376088,
			PacketsReceivedControl:        2376138,
			BytesSentControl:              525691022,
			BytesReceivedControl:          227044870,
		}
		assert.Equal(t, expected, ci)
	}

	instanceinfo := func(t *testing.T) {
		t.Helper()
		ii, err := c.Server.InstanceInfo()
		if !assert.NoError(t, err) {
			return
		}
		expected := &Instance{
			DatabaseVersion:             26,
			FileTransferPort:            30033,
			MaxTotalDownloadBandwidth:   18446744073709551615,
			MaxTotalUploadBandwidth:     18446744073709551615,
			GuestServerQueryGroup:       1,
			ServerQueryFloodCommands:    50,
			ServerQueryFloodTime:        3,
			ServerQueryBanTime:          600,
			TemplateServerAdminGroup:    3,
			TemplateServerDefaultGroup:  5,
			TemplateChannelAdminGroup:   1,
			TemplateChannelDefaultGroup: 4,
			PermissionsVersion:          19,
			PendingConnectionsPerIP:     0,
		}
		assert.Equal(t, expected, ii)
	}

	channellist := func(t *testing.T) {
		t.Helper()
		channels, err := c.Server.ChannelList()
		if !assert.NoError(t, err) {
			return
		}

		expected := []*Channel{
			{
				ID:                   499,
				ParentID:             0,
				ChannelOrder:         0,
				ChannelName:          "Default Channel",
				TotalClients:         1,
				NeededSubscribePower: 0,
			},
		}

		assert.Equal(t, expected, channels)
	}

	clientlist := func(t *testing.T) {
		t.Helper()
		clients, err := c.Server.ClientList()
		if !assert.NoError(t, err) {
			return
		}

		expected := []*OnlineClient{
			{
				ID:         42087,
				ChannelID:  39,
				DatabaseID: 19,
				Nickname:   "bdeb1337",
				Type:       0,
			},
		}

		assert.Equal(t, expected, clients)
	}

	clientlistextended := func(t *testing.T) {
		t.Helper()
		clientz, err := c.Server.ClientList(ClientListFull)
		if !assert.NoError(t, err) {
			return
		}

		// helper variables & functions for pointers
		falseP := false
		trueP := true
		stringptr := func(s string) *string {
			return &s
		}
		intptr := func(i int) *int {
			return &i
		}

		expected := []*OnlineClient{
			{
				ID:          42087,
				ChannelID:   39,
				DatabaseID:  19,
				Nickname:    "bdeb1337",
				Type:        0,
				Away:        true,
				AwayMessage: "afk",
				OnlineClientExt: &OnlineClientExt{
					UniqueIdentifier: stringptr("DZhdQU58qyooEK4Fr8Ly738hEmc="),
					OnlineClientVoice: &OnlineClientVoice{
						FlagTalking:        &falseP,
						InputMuted:         &falseP,
						OutputMuted:        &falseP,
						InputHardware:      &trueP,
						OutputHardware:     &trueP,
						TalkPower:          intptr(75),
						IsTalker:           &falseP,
						IsPrioritySpeaker:  &falseP,
						IsRecording:        &falseP,
						IsChannelCommander: &falseP,
					},
					OnlineClientTimes: &OnlineClientTimes{
						IdleTime:      intptr(1280228),
						Created:       intptr(1661793049),
						LastConnected: intptr(1691527133),
					},
					OnlineClientGroups: &OnlineClientGroups{
						ChannelGroupID:                 intptr(8),
						ChannelGroupInheritedChannelID: intptr(39),
						ServerGroups:                   &[]int{6, 8},
					},
					OnlineClientInfo: &OnlineClientInfo{
						Version:  stringptr("3.6.1 [Build: 1690193193]"),
						Platform: stringptr("OS X"),
					},
					IconID:  intptr(0),
					Country: stringptr("BE"),
					IP:      stringptr("1.3.3.7"),
					Badges:  stringptr(""),
				},
			},
		}

		assert.Equal(t, expected, clientz)
	}

	clientdblist := func(t *testing.T) {
		t.Helper()
		clients, err := c.Server.ClientDBList()
		if !assert.NoError(t, err) {
			return
		}

		expected := []*DBClient{
			{
				ID:               7,
				UniqueIdentifier: "DZhdQU58qyooEK4Fr8Ly738hEmc=",
				Nickname:         "MuhChy",
				Created:          time.Unix(1259147468, 0),
				LastConnected:    time.Unix(1259421233, 0),
			},
		}

		assert.Equal(t, expected, clients)
	}

	serversnapshotcreate := func(t *testing.T) {
		t.Helper()
		snapshot, err := c.Server.SnapshotCreate("test")
		if !assert.NoError(t, err) {
			return
		}
		excepted := &Snapshot{
			Version: 3,
			Data:    `KLUv\/aTFeAEAjeAAOuOELE2wkhEbvGpFNp3\/lL6F\/QtsvQL+1czMVGFGZESk9Io3xQ0eQdv35ihoaUFN+NNTCDtDIBmwWyMj4LXst++WmZmZmWZIrYrZKf0p\/SmdIboCtQLbAv2iNEl\/Rqd7cNZIc8m6Mo\/pGca0lt\/yTxkmEWXe37pcMOnJ56Z1VNqahbP+FN+adXFN41X2WEz2W4wXdS7KRtTDPJVt+bv1GtuctVDCGeF+bTEw3qnr6pJ19b5NiHE3X\/SSL2qStIUr\/frfmtHnTX8mvWTe+YVJ+rqozbuFnSvAPijKro4z37yquJayf8GPaVJM59nFM+xFn62nuCbpz9Jn1\/7J4lJF48Nc1WfRV7zPqFubZHFMYZWLKNwtfinVEb9qN6R4snn+nCZZNcRRDyGjhLC+KL+GwNU3HPVqh7Bc1AnhIiuEsziKEcI9Au6OXoZrIaBzMMIwHCL7\/VmIzTsFcLqLViHBKSV9Ecb36HPwO9c6Fu580wVQmgZNJyAAnAhAzrQRLS45w69J38EVH9FSXbmATXM8FQA7Fe+MaBnRYoCG9e4SydgdKtuk7C0LT845+Ca2OSehe5FL19yjLTxJFO5XlzS+KMtzaWlavNLYRR1MXeotC7vkF2HblXWkTm14vx7RokWv4ekpvt8NHR8UaKDp2uWiVZwUBVYkUJLygmVsE4mCBUnnQpQKJqKcwHBOXKTAsCCJfKWky53CcTHhdMjCTeeCmp7NeVmBT7CgTEwVFAawiYkKx+HeHCvRYucFCQuTEpyYiVmZFCueksliYEVfJ1CBFiwnVPKKifkyTqCUSIFRYmpZBoOTYTgmOKdjosPhfIdDwilBiTKOsZh\/iV5flWRdvGMxWbZeyq8vWiOU9r76Iv0df5Z1dnk0DWMwRlq3nlNZBq4oCMGHDWz4MIDWckCNHohI8EMJFvggAAk5AHFRFOYlou9gTfvqqytT7\/kggx5B6CgyowMCmxoR2PiQoCb6ClsjEWfrW0kMipcVCnDrZeiUFis0SotIYWWgsKLs+ebdes2l32Jwa5Oyh8leo\/DWMw4T3KL0CJMyUUoxAPLJQeRm1v1UO\/1xfpVQpnuCHJavVgjGSyc31cmDhRoYIPghABFRnRD\/23HCTzWFdYKmGTOmmFqw48OCjZoJ42Q3hQ\/+qw\/CN1vI0klCd3ISekmlnvKVD9Sg8bEBDZVTxwm+Xac3IzTtOauE4hQdlG101+38XoSvhhtsaOhxI1+Wlrrpvaf3hGLUc04RQtFuvV\/9CB74IYQWbi4QMtNBDDQ9drC5gcjHRhEgICih5oMiNZlrG8oQITMcQHCTw4QbHD6gwUDIjAki3NzQ4BPEBq0Ux\/nu89voXkJO\/9VY2qkjBjQ+Qyr46CBCE4SHHR7U6FADTYzysaTZ1g1tnW9920JPP3mjnReHGjQ\/OKiJ9+mXd352gq6bZN0PQ\/FmerG0tMWF1unpvSlf66BpPgtdmyUELb2QtDFHVykcMmRGpQR0Dr8N8btpn47Q3nuC5tNJQ\/TJv\/tV52ANINzQGOHGhgygxfhOS9\/dOEPb0gmqlUpo0q5CEMb9dnyPPhnRQOMzwY0aFnT0CMIBl33GmJYbmDGixaVjjGfvYnH9nVecghbHm+sUZbwQ3dVCkUIJ6bsr9O7jTCXuFNbe2iSNvyi7esZbisnWd0zx6jULY2pnr7MrhuIV3g96UUznFJ5RdMFgFAnESOBFn5hkNle9yBhPMS5qZVakdfGO8SxrrzPHrv7CkJkaMHx47DATOYy+Ea3z0wxJWz+iaZjK0ndwdnVNchbvbBQWRR5dnZC8WGB7Uk5cUOaWSJF58ooMFpQYMiooy8oJGbyDYq8ocVErKR4XjkVJlmHhiCiRMXE7lUV3EtMkyyaqE7cCM2DVjgIu0hIVMHWBj9mcuvLVL+tfri51LJ7xV9h6ydyKUj7t4tHJaO\/U\/Mq0haMY\/P2p1xJWfX99VS\/ST03YeqlrnJq0V2a6q3KthsKSuFSL4qE2Zb7AckVHGNF2kPIirs5SQONa2ZaYHBaKYtYSKmVeiZCC9VGvnrhKE6h0xw6bFpNyMEuiQ1kY4WZLjpMnJImNNGD4lNAoI2NKFBMZX1wwTi8u9FTomSA8WWFR1LuCDxivxIRQThFxOvFexnGPBLtvxbLVBBQ+lcaYMIZLS+LGGVpLhD\/LLr9lXblrmFKJRjRJW2+fs5fO+1PE9mWoTnaCpLYSik839O\/s\/SepCRWWFSUoTEhMoKOmBRs1OACx4aGBTQk4bKyUnz\/POOveGXI4T1C91ELoyaofph9PEHKzgwc3PTa4+fFCDTVuZsDxQYHIzI0RZHxIpdyYVvgPOUgv9NNJjCH9sNvYTQcrsEHzgJAZC4B8PvhxU6ULEdAEsYFGx5CZF0e0WCAa9AgiCA8ZRAeniyulD0f3KMXQ3FpC02oNWazxjnf+KDLDg0iNDx1qaNxgo7VQYMHn0SE+PhVg8PGB43Mjg88MN2zmiJZ64uCUdN+ZnaWYUsg5GSNkaZ77afkAkM8OCWyA+PiMAEKNjyIzroUhNjYksLnR4xNkB5sSitBskIHNhSNa6IgWCR28MWQmCA8p+PhogMNMDDg+U7pHacbT3lNG6O2ZrYRmniRE30RxrJXFoyLFywutwmQBEVBzGGBR8vIsNhcXLDEuKoD0oAGigw2RB2x47PBB1tfPTVtr\/TFCn\/9Ce8oSetBGaD88oY7RPiBAcAgBogMc9dk36awPvq95Q5JSC719pYOmae1H6651D2Oc5fu4IazdhXB9r6GLpZbyrfPAkLicqMRKhWV7Ji\/VxEqLBDNpcTnJVFUUFgIENiNgQLPDkBoVdlz3Lr06fp3gky9LC8k4\/UtI22ffproDo1JJseKC5ERFyfUsJylYixRZCg\/Jk7CA0QIMn7v1c7jOPbWGrKQTdCvc2EJ6ihvfN6FJjEY9LzhySiCGCJL+IiNx4iKa9bZOGZ17Dkbq5K1URltdpTDaaN+U0kZnHYTw3vumjc\/dZ6171jn4oKTw0knldbZaeN+D5q2SOktvlfVK6VykNtpLH4XUQhfBFz3NR6qD1975XqRuQvtilHTKOkcshHBCOzKBjpoaQvjU2EFDJ3TTQQels1XGERNq1OQgMjH1mZTUwQhrrY7WOa2dk1IaIa1Tyvmcq5Fe5+C89FY4oY3RXQrvvXU++t51VF5XYb3TwTeDdiatOZ+hwSmlffEHReGqVilrdZLmyNEjCOqg1Vo0x7auorVsKhPCK62F3E01hww0MdSwgU46C6O001JaWZBo8WR9V84P3LPw5CZ01MUJZbXPWeum4466yeeIDx4+4PBjhrYe2zCNu3i0TeMV709p\/pnkuVyklz5bWAa2+IVZm9YYu2bFtqmSYrKr58+41jxXNlXY+kzzvnnnop73K8Mm3V8ujL\/Fqb0v6zUZMDzDoIp3jifLM1l4qq5sSy984pr2mdb4+1VZnit7mCvmJfOuufSra7\/F+Cz71zY8QZiDGe4YBrtN0t66qMy1xlL2U\/XVlFiMxjaJa1yC91kZbul80k3qIPTTtGmcv\/QvmXOS2hcK955dGq8krm3VZ37BaBt2uXTu4dVrmuKzelzHXksSx1nnu5Xh7MKq752LuvrWswvTOhbvfeursrbsu8Y7n0kxm\/fL+uyyOlyy+AWjc1EdbOEq5l84hvtTr1\/btMbY1TPcterq2qYtvqVJT7prnJrhOm+5COMv2ot+ozCM569tSVOkR2iA4VMCCDQj2gxJW4\/iykBdV+wzMZ+MJOGJK9A9GSj44Jedeyidu0JiN+jhUj0K0GwyrwU7aKqMNd4XeFVehvmoo5iTUimFzMzIAIAgBADj0gAUFooapKmehoHZAMOAQBAcA8ZAUBAUGgXGZyAgAABAgABAAIIABACEIAhCohwww9oApxt1YJlRFykMPOU2dtuqo3xKQWksdgwT49jtnjMl7fNOc0Xf4zcqLpOyhMohy3mtuZzbHD2ovVdzfRCZn\/uFA96acPjLOYsevfcxJ1iw6C7uPQtx2IVCOscIt\/8it\/mTEZutFGgudjgWwFwkuVmAXyywBWNx\/I5dYIMNQ0iLU4URMmcxXUH+WjNgeflXoTNEIsDBLCatdPg7ceFt8qQl7ufJ9wyB\/Pm6bfgPuzRgnU1YfItHQrPEP8bfVECqTwLQuGSSq1t\/QLkadd3vl7kPcW7iVaZ91yCdA3B2kYqA\/oqoyPlAnn8TAFo5FeErUwynb96jSya6SoYKF8mjhKiF\/YJJCP6k7oRQOcHix9o2Quw0Zbl5yQbzraYZrm93gfU+TdPWZLqRID8QrQpBYXTUAamRRhgKJkRu1ioQlowEP56lYNmTt9sKRqkDtFiYXQji7GmNH36xkOGLiqbg88UvYCNh9lQK9kKwzYAF76RK07VyE7xSBBiCnxK10bMEzf9ny5uySqX99WitBlIh45G2CeYDBfY0k7SMMiU3kaQL08zoWelm5ARQrj2mxpxNnBG1HmZuT0cJjvD+7zSpVElFevbI1dzKkHW80Tw6AlpKh5NtlUbakx7Yl9gDVHWl9C79pNxbiDC9FJKBuEuPWHUuhy9Hnj2FHMHwCBUh5pu+GOWymWAqbMnm8tIaWyJFeGa0ubTJfNGObXOuMDgpVvXJjOrTXWLzKF5NVWcktY7kuVS4n1UbsPVWqglU0Re9mOsI7OMgMGYEWqlbhq77+FZTguOoSZQx0\/lOtosBvfibQh+VqNFobqn94F7qE\/frojFfQXq0K9uriRemqDpspkLu1qFiINZXJUtvQ16x\/pHOlPBHcKCQJiFeS5RZEDRwOM6eQIyQ\/WZKE8+avxtSl6NuYo1O87aqrGLCozEYVZCtiHC9FhF84MsuFerWDtO7Ptxh3RVXr3B8huJCMRcU\/BN\/70PMXzR6KN2yeEVznZfyyrFfXlHajiO6xF+k74NUaDbmyg0Z4sGfFgVkdA+p64XZO+Z+42td31jL328VWbQv8oodsV4T6ODQrn0vFVAHsLWiPI\/14O88MItGDEBF6diAs+HuTDLgDgv7REvA9Plm5M6f1KcmLJWxB618JyyMCiMPhQHsc1BBsY+Hyv9COuGDphWkbbB9zsEGU\/RU2nrcve0IaNwTuMBKluBpJrWMIpxvwFQCcijbcMAKRbmgR+NjDl0f0DOsuc6rpws6+hi9PjL19NYgWK8Ttd+TM6KVAKiiq81\/ZPpBR1pbh6Rdft1IRpuXzkem9wz8T3EvInqIx0eIkMXoWzdbCdkv8l+fCCR7El9e39YHLxFj6n+2UJ2nyhVV8miQ8SCeGi0V3HfgK9sojWFitAOWpIvX4XpWkskSy8CbhoJ2g9sNk3s6zX+tie2Hu\/V914A\/wZxJ31xHM3ApY+ZlEsCN+8MQPcyxuqEUpkpq5nJ6UcWg1Jc1c8CVJf+QBqlydevWcsn\/iHxcz22B+Rf7bN9o0dQLNlDfY8GfJbibAKGKA197\/oRlZ0WFF5+rWOyV000nOVJRAPbFp0QpSvlfgrJSSpVDDPaNd0qcQidFKp3rN7HC9asmmYzwygSMX4SASaFWNuAZg1XSA6SwUGnmljAAfYu4DjQ5UkQYEZBIwrOmHLIs7BNDQk3p14vD4\/4gA3pBW1mgz8dktPlWm\/zSg1pM+QKfMwiuIe1Qn4yoTt+gg95UaN4y2IyjajT2Vj3vBpli+zLsnaYKcjh8rz0G7GnenNU4A+NM7hHAEEdG+2WsyRiej9dGoXLIHEwiNRVqfjHkrgP2FnZlPKhFAUVNWI2QAwsuliuLxuLgKJN8wvySwTuEBwl+wUarUZ7lCvQPc98f0qnn9JD2LDx2e6YOe4u11j6sQRJChsX6OzIJMPDv36K4MGEWTydVArgC0GJ8vG0fKsgCf4NuCOWfOFdw3nHZqzNxByU\/OvfbRDZgvg3m5413Kx9C\/5XxLwL\/GlxSUO754afz+hHwe\/r8Rfp6AO7cBhxAZI7CM1Cag0DkG5VPmwFpAjYxJiOTA5r603UgDgB7hYT1I2ic\/ONPxoT+8+d2vwsQw8b0D41pPB6B8W3dpG3P+IM1Ut\/pb8o788klBmrvD3biX9im9PfXNNlcoJ34mTctcc1fXIIk8o\/qRvBniheJ7czQviM4nkAHGPvzpNZd37rPebGOZO5eTuvTy6SbGzMGCmQ+LuozvzhievL6bjWJ3PfogI7P78f7i19R13glvqFMr+HIh9CY+MMCj95ssN\/Fn2jhW\/+SgXmbf4Lf7q3lxOwVT6Y3bvzlRvF2+zWKsW0yMMtem\/sFhJw9TVv7qx8\/7Oxm3QlyvTor1\/PB8aD1fWi6J\/0w1bmmtURqk1ag3K2caQNEOHp3\/6fbo2eH6VgvN31CM38P+08vPXfpDHvt8zcHRFbNmw5WZoPlhQpTH5IvvmyVQ\/\/jD868OPBj+UW2bzgokCejY+34VvAUBRtvoGpMJyF9lhzZKyUrqZIXSyStfZBR+tCQch+USW8HbtbZToZTz37ALGTdYkXxdm52woAREh0PQ2wRxrZ2CY7UWkYGoXiP9Ew9sZz2Gkeaetee0IiMtPC\/3wREmbGLOChh3e5JPDFJFNw1zKSoBPsGDxJLmp1HHmIzjyMPbGMvsKNJmCJyLtsmFWISGn3qga3wwzC4k70NpuUJbDPCsAocVYYyDP+nlNGzAoj9KAycl6i00PIFXysjDZiz2B9KGJgLQ2CvZVDb+EAEFhGA31xOYEtdGuViADl8EmWXYdT0Jfpr6xf11dXHQv5zuSegw21WOyIjjaLd3t3DMm2C6kFAI+OTnjOleJdWOsBWgH1f4oFkLswvkhFZdkVM03FVUE6hId0pH8xepENyWjMmRNx5+Bh1HbhshE3MP5UR45gxfhG5CL7wX2SI5wlmK32TABSwtBkHUxUy9UWhUCeWQB8BJDAwKHvEL4YxG4\/xOkyQwXpUzGFSAR2DE9cnYjUzhnCTq\/CXFp2XOTuNI\/i5zQsxvABSmfcpqqQ66IZ\/R30CZ\/ycEukjsTSweUWXyZ6YCljLtcdrxcK2UWBTg2pJVqh0vTFy9HUK72UA1yBui1IjOCltTDRS8uFqC+aHWIegrdeTqBxImCK\/dYzfZkVZRHB+43+kL1Gpi0MUBKTxgPBBlKgQmKT8Z2Kl+Sir38Q1mJM+jiPOpTNfUKfvD7GAyVZPUCkDGFN6NBnZfJdVy+JShf9MUA4xrDUwoMOuEZ2K\/CcpmWEIaCTQGM5Rj8U0Tthj\/iRFYbrP8w9Z+bLZAJUIeWsP4hCV9NCipGBjqoeQuGmU+O6E+SkXzjaTW07wbNg6uwcWMOl7d2jWUDUPA2Mn\/AV4MPP6Ibm95QeGKnsu\/1bBjRRNlQlCNZIZYOwcHLKmMz\/iW2kp6oX\/aNcuncn7MlLPnCn3mjJ5zOiBaJeLjQFscIFFwjuN5biU+JJEKJvLkgkG6\/I0r2UQD+oJ4sy\/qkQ\/SysQAc0JxZwiPgPLg1YNDmvhAOmw8Iz3cFgW4yEyqmOcCfQcCRRmm+kmlIyFEc+fFzFITMkHIPCQwEMiJozrIBIR73cyoJ697TtIKvF+KCC4MfMXDx\/b+4euzuN+YyUDBYzK+9YW3HyeGi1V0ZDH\/nCPeFCh7SdziEhb1b1RSYnmCFhiJfzp7UmUnvH8iJKArW4msJUAyIdQVAU4yr0p6SrbDcc8DANAEjTlqJEdZG02RRQGLK5y5Pknjzycs9TXDrqcGVHqosNKY8umxZaDaHG2IMHbC6buu\/rL1CesCzg534okAXrJoa\/cxHMaMEpbXYalnZyuZuC5wiaUB3Is0wxWmtC61L5IvOtesQxJFHrTFB3jQBVIDCYsIC5yRZVpiU2RO0HKwhhVVMnLZGHqOFXGjiJ4EY6bnBM0\/U\/XhQkwFLi9pqHFnrZgNjxCVYg7NPLcAVMYfTHXGwZViDxojozzizwGgADoei6TJ+Zf0IGKcWwDRux1AEf6QgzUsFz1rwggzDyZytjzkxu6GzI1XKjkG3IYHRZs2joOdoJSBqhJDnjoga2U9oMrfgubriJcI9V2aEC+LNFDkiEJhYuI43NMEgQGzywwqWojyYhkTCvqYiAebcqCrv0zMnEL+XL\/JgMG0TJiMghnu6FSQ+PqbhEdMsLcBkNg+Z7jLSfBqDDZchQaP61s0BQBQKvj1ugFK1JB1IbDLAS3k4LIg9usUsCHPGaWt+5aAC6O0asc9CarOXjl5hbEM41d\/lWRoQTKe9KQMSUXQiyzB1fx+XBDEvZOC2QEFtWebrxWABr1wPz41kham2epuhIwBI7HDMn3LtKhE1Jro1nWtZBBGuOLCCt9OqfdsK0WKy2D5pNe9GSha5FLKvHAM82oq1X53ck92SMXkmqg4Gr3cdiCNQHWps2m0s7hKWtLxjkkG1v5FBSNLizOUCp+ExjyjeisB3NFFBFBzNt1vxs8HIrFPccaShG2Nq5vTlwc+LfYHFGeiuqHkMKTnC7G8FRynfdzQEo0RNYpnJcsFPVrWRGB8nKifWwcYnP6VrW1+ziHaVNMQVH1I7JTUtBNNojl3DMsD4ZtrRgkiqul7CrQ1uEbkGC6hJDW1HxyOk+UjPgwXkT6TPsnMnerbVaDm2KecFGoRqI3kMKdozq6+VZ0HVi1fgbgnYoodOseijhO91ZH1TboqCYGjABstnOHNzJKjw9tGdxr0iS0DTwxAoQQ6VPBmm0sQpY7RFm8wPZGKej+ukqzdYyoce+2tgF7WiLiTw8K9IpWzf0pl2umb13OG6B714ynLuma4tacQAvIxR1O2TX6i7G0VxLPviVRcg0KMhzWkIk7QNU5sRqBmaSB0vQjuHSis6FbSdGJVesaOwlAJ3dZg7LtKOAYMu28Wqc5KB1nLpNTM9FXCyd0IFhbf9bs8qfFKSbDaEEPLFJpD+60Jn8cuKgihX+B6tbg3BiYHhadZy3NbnMozUmr7f22CVhatA2ZxcTUsShYOEIuyPWRQtdi5ayqORz5HF1C16NO55rgboChvYMQxRiBlr4b8uWU6YWnyp++w51f4o+dGzqORnhTs+1uRDyVBslITwCNJ3qIZmEYbWCPIr+nCUbo62CjGfVDdoDnAzDkPxNdrd571cjm6AAdzAB+JUyOwEq0sjywUWfBNtGNyQuOZqAnhdiz+qxTVY+qXtOU7T6SOXHyQewK4FinnJc5xIIMUNl6A9GWn6pQ0J1TmpiMSKVF1cStwv0UsaGor6S\/qDm9MkSY57hvEMjOyOTsvmzewdGjnOZLis1g7VqDfgl9ENRC96UIdL1Qq4J9iJ6XyaH3Ns7nhc3qvybh5JJC5vxz0b3O8xK9TY8vEldL+8WRBuyG3tSRGaf\/BBNqlWnTiafI6cVo8jgQCZqCipjQOgzRH9TaO\/wSRcUXr+ARRm2kIwEMkFZVC9nUfnSAOcPWo4AZ0dzcP+CQiwAaDW5YKExBEMYMGfhv1fXsej95BciqNHqeKq3zhFJb8rQDgJj5C3T\/Nbq\/kFqOPvBa+Wg\/InrjVSNdOyRNN4wWXG54aa1pkR+fF6qcCYFxghqJSjqK2bqHqKZkRIuGJDzO++gr9GJo6tjbFLpCtDzWJoZjaAnBjvTjWJIdbWN3h07GBN9UAe6H31wgePJV2\/yb6dZxyn8+LhroyMYP2yxWLw==`,
		}
		assert.Equal(t, excepted, snapshot)
	}
	tests := []struct {
		name string
		f    func(t *testing.T)
	}{
		{"list", list},
		{"idgetbyport", idgetbyport},
		{"info", info},
		{"create", create},
		{"edit", edit},
		{"del", del},
		{"start", start},
		{"stop", stop},
		{"grouplist", grouplist},
		{"privilegekeylist", privilegekeylist},
		{"privilegekeyadd", privilegekeyadd},
		{"serverrequestconnectioninfo", serverrequestconnectioninfo},
		{"instanceinfo", instanceinfo},
		{"channellist", channellist},
		{"clientlist", clientlist},
		{"clientlistextended", clientlistextended},
		{"clientdblist", clientdblist},
		{"serversnapshotcreate", serversnapshotcreate},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.f)
	}
}
