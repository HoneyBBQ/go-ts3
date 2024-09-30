package ts3

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const (
	cmdQuit = "quit"
	banner  = `Welcome to the TeamSpeak 3 ServerQuery interface, type "help" for a list of commands and "help <command>" for information on a specific command.`

	errUnknownCmd = `error id=256 msg=command\snot\sfound`
	errOK         = `error id=0 msg=ok`

	// only used for testing.
	sshPrivateServerKey = "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAaAAAABNlY2RzYS\n1zaGEyLW5pc3RwMjU2AAAACG5pc3RwMjU2AAAAQQRamQdnvjuFVMSN3wpq246IZxO9kS0y\n0f54xgj47XwyPUvhbpk27Ot6Z6CkqvLnj05pNQK6j7XJPkVoym16tiSLAAAAsOwJzensCc\n3pAAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFqZB2e+O4VUxI3f\nCmrbjohnE72RLTLR/njGCPjtfDI9S+FumTbs63pnoKSq8uePTmk1ArqPtck+RWjKbXq2JI\nsAAAAhAIVVOJZP3A2+tO26RnAXBAaD6aPpDfr1QgoeFz2Rd7E2AAAAFmZlcmRpbmFuZEBG\nZXJkaW5hbmQtUEMB\n-----END OPENSSH PRIVATE KEY-----"
)

var commands = map[string]string{
	"version":                     "version=3.0.12.2 build=1455547898 platform=FreeBSD",
	"login":                       "",
	"logout":                      "",
	"use":                         "",
	"serverlist":                  `virtualserver_id=1 virtualserver_port=10677 virtualserver_status=online virtualserver_clientsonline=1 virtualserver_queryclientsonline=1 virtualserver_maxclients=35 virtualserver_uptime=12345025 virtualserver_name=Server\s#1 virtualserver_autostart=1 virtualserver_machine_id=1 virtualserver_unique_identifier=uniq1|virtualserver_id=2 virtualserver_port=10617 virtualserver_status=online virtualserver_clientsonline=3 virtualserver_queryclientsonline=2 virtualserver_maxclients=10 virtualserver_uptime=3165117 virtualserver_name=Server\s#2 virtualserver_autostart=1 virtualserver_machine_id=1 virtualserver_unique_identifier=uniq2`,
	"serverinfo":                  `virtualserver_antiflood_points_needed_command_block=150 virtualserver_antiflood_points_needed_ip_block=250 virtualserver_antiflood_points_tick_reduce=5 virtualserver_channel_temp_delete_delay_default=0 virtualserver_codec_encryption_mode=0 virtualserver_complain_autoban_count=5 virtualserver_complain_autoban_time=1200 virtualserver_complain_remove_time=3600 virtualserver_created=0 virtualserver_default_channel_admin_group=1 virtualserver_default_channel_group=4 virtualserver_default_server_group=5 virtualserver_download_quota=18446744073709551615 virtualserver_filebase=files virtualserver_flag_password=0 virtualserver_hostbanner_gfx_interval=0 virtualserver_hostbanner_gfx_url virtualserver_hostbanner_mode=0 virtualserver_hostbanner_url virtualserver_hostbutton_gfx_url virtualserver_hostbutton_tooltip=Multiplay\sGame\sServers virtualserver_hostbutton_url=http:\/\/www.multiplaygameservers.com virtualserver_hostmessage virtualserver_hostmessage_mode=0 virtualserver_icon_id=0 virtualserver_log_channel=0 virtualserver_log_client=0 virtualserver_log_filetransfer=0 virtualserver_log_permissions=1 virtualserver_log_query=0 virtualserver_log_server=0 virtualserver_max_download_total_bandwidth=18446744073709551615 virtualserver_max_upload_total_bandwidth=18446744073709551615 virtualserver_maxclients=32 virtualserver_min_android_version=0 virtualserver_min_client_version=0 virtualserver_min_clients_in_channel_before_forced_silence=100 virtualserver_min_ios_version=0 virtualserver_name=Test\sServer virtualserver_name_phonetic virtualserver_needed_identity_security_level=8 virtualserver_password virtualserver_priority_speaker_dimm_modificator=-18.0000 virtualserver_reserved_slots=0 virtualserver_status=template virtualserver_unique_identifier virtualserver_upload_quota=18446744073709551615 virtualserver_weblist_enabled=1 virtualserver_welcomemessage=Welcome\sto\sTeamSpeak,\scheck\s[URL]www.teamspeak.com[\/URL]\sfor\slatest\sinfos.`,
	"servercreate":                `sid=2 virtualserver_port=9988 token=eKnFZQ9EK7G7MhtuQB6+N2B1PNZZ6OZL3ycDp2OW`,
	"serveridgetbyport":           `server_id=1`,
	"servergrouplist":             `sgid=1 name=Guest\sServer\sQuery type=2 iconid=0 savedb=0 sortid=0 namemode=0 n_modifyp=0 n_member_addp=0 n_member_removep=0|sgid=2 name=Admin\sServer\sQuery type=2 iconid=500 savedb=1 sortid=0 namemode=0 n_modifyp=100 n_member_addp=100 n_member_removep=100`,
	"privilegekeylist":            `token=zTfamFVhiMEzhTl49KrOVYaMilHPgQEBQOJFh6qX token_type=0 token_id1=17395 token_id2=0 token_created=1499948005 token_description`,
	"privilegekeyadd":             `token=zTfamFVhiMEzhTl49KrOVYaMilHPgQEBQOJFh6qX`,
	"serverdelete":                "",
	"serverstop":                  "",
	"serverstart":                 "",
	"serveredit":                  "",
	"instanceinfo":                "serverinstance_database_version=26 serverinstance_filetransfer_port=30033 serverinstance_max_download_total_bandwidth=18446744073709551615 serverinstance_max_upload_total_bandwidth=18446744073709551615 serverinstance_guest_serverquery_group=1 serverinstance_serverquery_flood_commands=50 serverinstance_serverquery_flood_time=3 serverinstance_serverquery_ban_time=600 serverinstance_template_serveradmin_group=3 serverinstance_template_serverdefault_group=5 serverinstance_template_channeladmin_group=1 serverinstance_template_channeldefault_group=4 serverinstance_permissions_version=19 serverinstance_pending_connections_per_ip=0",
	"serverrequestconnectioninfo": "connection_filetransfer_bandwidth_sent=0 connection_filetransfer_bandwidth_received=0 connection_filetransfer_bytes_sent_total=617 connection_filetransfer_bytes_received_total=0 connection_packets_sent_total=926413 connection_bytes_sent_total=92911395 connection_packets_received_total=650335 connection_bytes_received_total=61940731 connection_bandwidth_sent_last_second_total=0 connection_bandwidth_sent_last_minute_total=0 connection_bandwidth_received_last_second_total=0 connection_bandwidth_received_last_minute_total=0 connection_connected_time=49408 connection_packetloss_total=0.0000 connection_ping=0.0000 connection_packets_sent_speech=320432180 connection_bytes_sent_speech=43805818511 connection_packets_received_speech=174885295 connection_bytes_received_speech=24127808273 connection_packets_sent_keepalive=55230363 connection_bytes_sent_keepalive=2264444883 connection_packets_received_keepalive=55149547 connection_bytes_received_keepalive=2316390993 connection_packets_sent_control=2376088 connection_bytes_sent_control=525691022 connection_packets_received_control=2376138 connection_bytes_received_control=227044870",
	"channellist":                 "cid=499 pid=0 channel_order=0 channel_name=Default\\sChannel total_clients=1 channel_needed_subscribe_power=0",
	"clientlist":                  `clid=42087 cid=39 client_database_id=19 client_nickname=bdeb1337 client_type=0 client_away=0 client_away_message`,
	"clientlist -uid -away -voice -times -groups -info -icon -country -ip -badges": `clid=42087 cid=39 client_database_id=19 client_nickname=bdeb1337 client_type=0 client_away=1 client_away_message=afk client_flag_talking=0 client_input_muted=0 client_output_muted=0 client_input_hardware=1 client_output_hardware=1 client_talk_power=75 client_is_talker=0 client_is_priority_speaker=0 client_is_recording=0 client_is_channel_commander=0 client_unique_identifier=DZhdQU58qyooEK4Fr8Ly738hEmc= client_servergroups=6,8 client_channel_group_id=8 client_channel_group_inherited_channel_id=39 client_version=3.6.1\s[Build:\s1690193193] client_platform=OS\sX client_idle_time=1280228 client_created=1661793049 client_lastconnected=1691527133 client_icon_id=0 client_country=BE connection_client_ip=1.3.3.7 client_badges`,
	"clientdblist":         "cldbid=7 client_unique_identifier=DZhdQU58qyooEK4Fr8Ly738hEmc= client_nickname=MuhChy client_created=1259147468 client_lastconnected=1259421233",
	"whoami":               "virtualserver_status=online virtualserver_id=18 virtualserver_unique_identifier=gNITtWtKs9+Uh3L4LKv8\\/YHsn5c= virtualserver_port=9987 client_id=94 client_channel_id=432 client_nickname=serveradmin\\sfrom\\s127.0.0.1:49725 client_database_id=1 client_login_name=serveradmin client_unique_identifier=serveradmin client_origin_server_id=0",
	"serversnapshotcreate": `version=3 data=KLUv\/aTFeAEAjeAAOuOELE2wkhEbvGpFNp3\/lL6F\/QtsvQL+1czMVGFGZESk9Io3xQ0eQdv35ihoaUFN+NNTCDtDIBmwWyMj4LXst++WmZmZmWZIrYrZKf0p\/SmdIboCtQLbAv2iNEl\/Rqd7cNZIc8m6Mo\/pGca0lt\/yTxkmEWXe37pcMOnJ56Z1VNqahbP+FN+adXFN41X2WEz2W4wXdS7KRtTDPJVt+bv1GtuctVDCGeF+bTEw3qnr6pJ19b5NiHE3X\/SSL2qStIUr\/frfmtHnTX8mvWTe+YVJ+rqozbuFnSvAPijKro4z37yquJayf8GPaVJM59nFM+xFn62nuCbpz9Jn1\/7J4lJF48Nc1WfRV7zPqFubZHFMYZWLKNwtfinVEb9qN6R4snn+nCZZNcRRDyGjhLC+KL+GwNU3HPVqh7Bc1AnhIiuEsziKEcI9Au6OXoZrIaBzMMIwHCL7\/VmIzTsFcLqLViHBKSV9Ecb36HPwO9c6Fu580wVQmgZNJyAAnAhAzrQRLS45w69J38EVH9FSXbmATXM8FQA7Fe+MaBnRYoCG9e4SydgdKtuk7C0LT845+Ca2OSehe5FL19yjLTxJFO5XlzS+KMtzaWlavNLYRR1MXeotC7vkF2HblXWkTm14vx7RokWv4ekpvt8NHR8UaKDp2uWiVZwUBVYkUJLygmVsE4mCBUnnQpQKJqKcwHBOXKTAsCCJfKWky53CcTHhdMjCTeeCmp7NeVmBT7CgTEwVFAawiYkKx+HeHCvRYucFCQuTEpyYiVmZFCueksliYEVfJ1CBFiwnVPKKifkyTqCUSIFRYmpZBoOTYTgmOKdjosPhfIdDwilBiTKOsZh\/iV5flWRdvGMxWbZeyq8vWiOU9r76Iv0df5Z1dnk0DWMwRlq3nlNZBq4oCMGHDWz4MIDWckCNHohI8EMJFvggAAk5AHFRFOYlou9gTfvqqytT7\/kggx5B6CgyowMCmxoR2PiQoCb6ClsjEWfrW0kMipcVCnDrZeiUFis0SotIYWWgsKLs+ebdes2l32Jwa5Oyh8leo\/DWMw4T3KL0CJMyUUoxAPLJQeRm1v1UO\/1xfpVQpnuCHJavVgjGSyc31cmDhRoYIPghABFRnRD\/23HCTzWFdYKmGTOmmFqw48OCjZoJ42Q3hQ\/+qw\/CN1vI0klCd3ISekmlnvKVD9Sg8bEBDZVTxwm+Xac3IzTtOauE4hQdlG101+38XoSvhhtsaOhxI1+Wlrrpvaf3hGLUc04RQtFuvV\/9CB74IYQWbi4QMtNBDDQ9drC5gcjHRhEgICih5oMiNZlrG8oQITMcQHCTw4QbHD6gwUDIjAki3NzQ4BPEBq0Ux\/nu89voXkJO\/9VY2qkjBjQ+Qyr46CBCE4SHHR7U6FADTYzysaTZ1g1tnW9920JPP3mjnReHGjQ\/OKiJ9+mXd352gq6bZN0PQ\/FmerG0tMWF1unpvSlf66BpPgtdmyUELb2QtDFHVykcMmRGpQR0Dr8N8btpn47Q3nuC5tNJQ\/TJv\/tV52ANINzQGOHGhgygxfhOS9\/dOEPb0gmqlUpo0q5CEMb9dnyPPhnRQOMzwY0aFnT0CMIBl33GmJYbmDGixaVjjGfvYnH9nVecghbHm+sUZbwQ3dVCkUIJ6bsr9O7jTCXuFNbe2iSNvyi7esZbisnWd0zx6jULY2pnr7MrhuIV3g96UUznFJ5RdMFgFAnESOBFn5hkNle9yBhPMS5qZVakdfGO8SxrrzPHrv7CkJkaMHx47DATOYy+Ea3z0wxJWz+iaZjK0ndwdnVNchbvbBQWRR5dnZC8WGB7Uk5cUOaWSJF58ooMFpQYMiooy8oJGbyDYq8ocVErKR4XjkVJlmHhiCiRMXE7lUV3EtMkyyaqE7cCM2DVjgIu0hIVMHWBj9mcuvLVL+tfri51LJ7xV9h6ydyKUj7t4tHJaO\/U\/Mq0haMY\/P2p1xJWfX99VS\/ST03YeqlrnJq0V2a6q3KthsKSuFSL4qE2Zb7AckVHGNF2kPIirs5SQONa2ZaYHBaKYtYSKmVeiZCC9VGvnrhKE6h0xw6bFpNyMEuiQ1kY4WZLjpMnJImNNGD4lNAoI2NKFBMZX1wwTi8u9FTomSA8WWFR1LuCDxivxIRQThFxOvFexnGPBLtvxbLVBBQ+lcaYMIZLS+LGGVpLhD\/LLr9lXblrmFKJRjRJW2+fs5fO+1PE9mWoTnaCpLYSik839O\/s\/SepCRWWFSUoTEhMoKOmBRs1OACx4aGBTQk4bKyUnz\/POOveGXI4T1C91ELoyaofph9PEHKzgwc3PTa4+fFCDTVuZsDxQYHIzI0RZHxIpdyYVvgPOUgv9NNJjCH9sNvYTQcrsEHzgJAZC4B8PvhxU6ULEdAEsYFGx5CZF0e0WCAa9AgiCA8ZRAeniyulD0f3KMXQ3FpC02oNWazxjnf+KDLDg0iNDx1qaNxgo7VQYMHn0SE+PhVg8PGB43Mjg88MN2zmiJZ64uCUdN+ZnaWYUsg5GSNkaZ77afkAkM8OCWyA+PiMAEKNjyIzroUhNjYksLnR4xNkB5sSitBskIHNhSNa6IgWCR28MWQmCA8p+PhogMNMDDg+U7pHacbT3lNG6O2ZrYRmniRE30RxrJXFoyLFywutwmQBEVBzGGBR8vIsNhcXLDEuKoD0oAGigw2RB2x47PBB1tfPTVtr\/TFCn\/9Ce8oSetBGaD88oY7RPiBAcAgBogMc9dk36awPvq95Q5JSC719pYOmae1H6651D2Oc5fu4IazdhXB9r6GLpZbyrfPAkLicqMRKhWV7Ji\/VxEqLBDNpcTnJVFUUFgIENiNgQLPDkBoVdlz3Lr06fp3gky9LC8k4\/UtI22ffproDo1JJseKC5ERFyfUsJylYixRZCg\/Jk7CA0QIMn7v1c7jOPbWGrKQTdCvc2EJ6ihvfN6FJjEY9LzhySiCGCJL+IiNx4iKa9bZOGZ17Dkbq5K1URltdpTDaaN+U0kZnHYTw3vumjc\/dZ6171jn4oKTw0knldbZaeN+D5q2SOktvlfVK6VykNtpLH4XUQhfBFz3NR6qD1975XqRuQvtilHTKOkcshHBCOzKBjpoaQvjU2EFDJ3TTQQels1XGERNq1OQgMjH1mZTUwQhrrY7WOa2dk1IaIa1Tyvmcq5Fe5+C89FY4oY3RXQrvvXU++t51VF5XYb3TwTeDdiatOZ+hwSmlffEHReGqVilrdZLmyNEjCOqg1Vo0x7auorVsKhPCK62F3E01hww0MdSwgU46C6O001JaWZBo8WR9V84P3LPw5CZ01MUJZbXPWeum4466yeeIDx4+4PBjhrYe2zCNu3i0TeMV709p\/pnkuVyklz5bWAa2+IVZm9YYu2bFtqmSYrKr58+41jxXNlXY+kzzvnnnop73K8Mm3V8ujL\/Fqb0v6zUZMDzDoIp3jifLM1l4qq5sSy984pr2mdb4+1VZnit7mCvmJfOuufSra7\/F+Cz71zY8QZiDGe4YBrtN0t66qMy1xlL2U\/XVlFiMxjaJa1yC91kZbul80k3qIPTTtGmcv\/QvmXOS2hcK955dGq8krm3VZ37BaBt2uXTu4dVrmuKzelzHXksSx1nnu5Xh7MKq752LuvrWswvTOhbvfeursrbsu8Y7n0kxm\/fL+uyyOlyy+AWjc1EdbOEq5l84hvtTr1\/btMbY1TPcterq2qYtvqVJT7prnJrhOm+5COMv2ot+ozCM569tSVOkR2iA4VMCCDQj2gxJW4\/iykBdV+wzMZ+MJOGJK9A9GSj44Jedeyidu0JiN+jhUj0K0GwyrwU7aKqMNd4XeFVehvmoo5iTUimFzMzIAIAgBADj0gAUFooapKmehoHZAMOAQBAcA8ZAUBAUGgXGZyAgAABAgABAAIIABACEIAhCohwww9oApxt1YJlRFykMPOU2dtuqo3xKQWksdgwT49jtnjMl7fNOc0Xf4zcqLpOyhMohy3mtuZzbHD2ovVdzfRCZn\/uFA96acPjLOYsevfcxJ1iw6C7uPQtx2IVCOscIt\/8it\/mTEZutFGgudjgWwFwkuVmAXyywBWNx\/I5dYIMNQ0iLU4URMmcxXUH+WjNgeflXoTNEIsDBLCatdPg7ceFt8qQl7ufJ9wyB\/Pm6bfgPuzRgnU1YfItHQrPEP8bfVECqTwLQuGSSq1t\/QLkadd3vl7kPcW7iVaZ91yCdA3B2kYqA\/oqoyPlAnn8TAFo5FeErUwynb96jSya6SoYKF8mjhKiF\/YJJCP6k7oRQOcHix9o2Quw0Zbl5yQbzraYZrm93gfU+TdPWZLqRID8QrQpBYXTUAamRRhgKJkRu1ioQlowEP56lYNmTt9sKRqkDtFiYXQji7GmNH36xkOGLiqbg88UvYCNh9lQK9kKwzYAF76RK07VyE7xSBBiCnxK10bMEzf9ny5uySqX99WitBlIh45G2CeYDBfY0k7SMMiU3kaQL08zoWelm5ARQrj2mxpxNnBG1HmZuT0cJjvD+7zSpVElFevbI1dzKkHW80Tw6AlpKh5NtlUbakx7Yl9gDVHWl9C79pNxbiDC9FJKBuEuPWHUuhy9Hnj2FHMHwCBUh5pu+GOWymWAqbMnm8tIaWyJFeGa0ubTJfNGObXOuMDgpVvXJjOrTXWLzKF5NVWcktY7kuVS4n1UbsPVWqglU0Re9mOsI7OMgMGYEWqlbhq77+FZTguOoSZQx0\/lOtosBvfibQh+VqNFobqn94F7qE\/frojFfQXq0K9uriRemqDpspkLu1qFiINZXJUtvQ16x\/pHOlPBHcKCQJiFeS5RZEDRwOM6eQIyQ\/WZKE8+avxtSl6NuYo1O87aqrGLCozEYVZCtiHC9FhF84MsuFerWDtO7Ptxh3RVXr3B8huJCMRcU\/BN\/70PMXzR6KN2yeEVznZfyyrFfXlHajiO6xF+k74NUaDbmyg0Z4sGfFgVkdA+p64XZO+Z+42td31jL328VWbQv8oodsV4T6ODQrn0vFVAHsLWiPI\/14O88MItGDEBF6diAs+HuTDLgDgv7REvA9Plm5M6f1KcmLJWxB618JyyMCiMPhQHsc1BBsY+Hyv9COuGDphWkbbB9zsEGU\/RU2nrcve0IaNwTuMBKluBpJrWMIpxvwFQCcijbcMAKRbmgR+NjDl0f0DOsuc6rpws6+hi9PjL19NYgWK8Ttd+TM6KVAKiiq81\/ZPpBR1pbh6Rdft1IRpuXzkem9wz8T3EvInqIx0eIkMXoWzdbCdkv8l+fCCR7El9e39YHLxFj6n+2UJ2nyhVV8miQ8SCeGi0V3HfgK9sojWFitAOWpIvX4XpWkskSy8CbhoJ2g9sNk3s6zX+tie2Hu\/V914A\/wZxJ31xHM3ApY+ZlEsCN+8MQPcyxuqEUpkpq5nJ6UcWg1Jc1c8CVJf+QBqlydevWcsn\/iHxcz22B+Rf7bN9o0dQLNlDfY8GfJbibAKGKA197\/oRlZ0WFF5+rWOyV000nOVJRAPbFp0QpSvlfgrJSSpVDDPaNd0qcQidFKp3rN7HC9asmmYzwygSMX4SASaFWNuAZg1XSA6SwUGnmljAAfYu4DjQ5UkQYEZBIwrOmHLIs7BNDQk3p14vD4\/4gA3pBW1mgz8dktPlWm\/zSg1pM+QKfMwiuIe1Qn4yoTt+gg95UaN4y2IyjajT2Vj3vBpli+zLsnaYKcjh8rz0G7GnenNU4A+NM7hHAEEdG+2WsyRiej9dGoXLIHEwiNRVqfjHkrgP2FnZlPKhFAUVNWI2QAwsuliuLxuLgKJN8wvySwTuEBwl+wUarUZ7lCvQPc98f0qnn9JD2LDx2e6YOe4u11j6sQRJChsX6OzIJMPDv36K4MGEWTydVArgC0GJ8vG0fKsgCf4NuCOWfOFdw3nHZqzNxByU\/OvfbRDZgvg3m5413Kx9C\/5XxLwL\/GlxSUO754afz+hHwe\/r8Rfp6AO7cBhxAZI7CM1Cag0DkG5VPmwFpAjYxJiOTA5r603UgDgB7hYT1I2ic\/ONPxoT+8+d2vwsQw8b0D41pPB6B8W3dpG3P+IM1Ut\/pb8o788klBmrvD3biX9im9PfXNNlcoJ34mTctcc1fXIIk8o\/qRvBniheJ7czQviM4nkAHGPvzpNZd37rPebGOZO5eTuvTy6SbGzMGCmQ+LuozvzhievL6bjWJ3PfogI7P78f7i19R13glvqFMr+HIh9CY+MMCj95ssN\/Fn2jhW\/+SgXmbf4Lf7q3lxOwVT6Y3bvzlRvF2+zWKsW0yMMtem\/sFhJw9TVv7qx8\/7Oxm3QlyvTor1\/PB8aD1fWi6J\/0w1bmmtURqk1ag3K2caQNEOHp3\/6fbo2eH6VgvN31CM38P+08vPXfpDHvt8zcHRFbNmw5WZoPlhQpTH5IvvmyVQ\/\/jD868OPBj+UW2bzgokCejY+34VvAUBRtvoGpMJyF9lhzZKyUrqZIXSyStfZBR+tCQch+USW8HbtbZToZTz37ALGTdYkXxdm52woAREh0PQ2wRxrZ2CY7UWkYGoXiP9Ew9sZz2Gkeaetee0IiMtPC\/3wREmbGLOChh3e5JPDFJFNw1zKSoBPsGDxJLmp1HHmIzjyMPbGMvsKNJmCJyLtsmFWISGn3qga3wwzC4k70NpuUJbDPCsAocVYYyDP+nlNGzAoj9KAycl6i00PIFXysjDZiz2B9KGJgLQ2CvZVDb+EAEFhGA31xOYEtdGuViADl8EmWXYdT0Jfpr6xf11dXHQv5zuSegw21WOyIjjaLd3t3DMm2C6kFAI+OTnjOleJdWOsBWgH1f4oFkLswvkhFZdkVM03FVUE6hId0pH8xepENyWjMmRNx5+Bh1HbhshE3MP5UR45gxfhG5CL7wX2SI5wlmK32TABSwtBkHUxUy9UWhUCeWQB8BJDAwKHvEL4YxG4\/xOkyQwXpUzGFSAR2DE9cnYjUzhnCTq\/CXFp2XOTuNI\/i5zQsxvABSmfcpqqQ66IZ\/R30CZ\/ycEukjsTSweUWXyZ6YCljLtcdrxcK2UWBTg2pJVqh0vTFy9HUK72UA1yBui1IjOCltTDRS8uFqC+aHWIegrdeTqBxImCK\/dYzfZkVZRHB+43+kL1Gpi0MUBKTxgPBBlKgQmKT8Z2Kl+Sir38Q1mJM+jiPOpTNfUKfvD7GAyVZPUCkDGFN6NBnZfJdVy+JShf9MUA4xrDUwoMOuEZ2K\/CcpmWEIaCTQGM5Rj8U0Tthj\/iRFYbrP8w9Z+bLZAJUIeWsP4hCV9NCipGBjqoeQuGmU+O6E+SkXzjaTW07wbNg6uwcWMOl7d2jWUDUPA2Mn\/AV4MPP6Ibm95QeGKnsu\/1bBjRRNlQlCNZIZYOwcHLKmMz\/iW2kp6oX\/aNcuncn7MlLPnCn3mjJ5zOiBaJeLjQFscIFFwjuN5biU+JJEKJvLkgkG6\/I0r2UQD+oJ4sy\/qkQ\/SysQAc0JxZwiPgPLg1YNDmvhAOmw8Iz3cFgW4yEyqmOcCfQcCRRmm+kmlIyFEc+fFzFITMkHIPCQwEMiJozrIBIR73cyoJ697TtIKvF+KCC4MfMXDx\/b+4euzuN+YyUDBYzK+9YW3HyeGi1V0ZDH\/nCPeFCh7SdziEhb1b1RSYnmCFhiJfzp7UmUnvH8iJKArW4msJUAyIdQVAU4yr0p6SrbDcc8DANAEjTlqJEdZG02RRQGLK5y5Pknjzycs9TXDrqcGVHqosNKY8umxZaDaHG2IMHbC6buu\/rL1CesCzg534okAXrJoa\/cxHMaMEpbXYalnZyuZuC5wiaUB3Is0wxWmtC61L5IvOtesQxJFHrTFB3jQBVIDCYsIC5yRZVpiU2RO0HKwhhVVMnLZGHqOFXGjiJ4EY6bnBM0\/U\/XhQkwFLi9pqHFnrZgNjxCVYg7NPLcAVMYfTHXGwZViDxojozzizwGgADoei6TJ+Zf0IGKcWwDRux1AEf6QgzUsFz1rwggzDyZytjzkxu6GzI1XKjkG3IYHRZs2joOdoJSBqhJDnjoga2U9oMrfgubriJcI9V2aEC+LNFDkiEJhYuI43NMEgQGzywwqWojyYhkTCvqYiAebcqCrv0zMnEL+XL\/JgMG0TJiMghnu6FSQ+PqbhEdMsLcBkNg+Z7jLSfBqDDZchQaP61s0BQBQKvj1ugFK1JB1IbDLAS3k4LIg9usUsCHPGaWt+5aAC6O0asc9CarOXjl5hbEM41d\/lWRoQTKe9KQMSUXQiyzB1fx+XBDEvZOC2QEFtWebrxWABr1wPz41kham2epuhIwBI7HDMn3LtKhE1Jro1nWtZBBGuOLCCt9OqfdsK0WKy2D5pNe9GSha5FLKvHAM82oq1X53ck92SMXkmqg4Gr3cdiCNQHWps2m0s7hKWtLxjkkG1v5FBSNLizOUCp+ExjyjeisB3NFFBFBzNt1vxs8HIrFPccaShG2Nq5vTlwc+LfYHFGeiuqHkMKTnC7G8FRynfdzQEo0RNYpnJcsFPVrWRGB8nKifWwcYnP6VrW1+ziHaVNMQVH1I7JTUtBNNojl3DMsD4ZtrRgkiqul7CrQ1uEbkGC6hJDW1HxyOk+UjPgwXkT6TPsnMnerbVaDm2KecFGoRqI3kMKdozq6+VZ0HVi1fgbgnYoodOseijhO91ZH1TboqCYGjABstnOHNzJKjw9tGdxr0iS0DTwxAoQQ6VPBmm0sQpY7RFm8wPZGKej+ukqzdYyoce+2tgF7WiLiTw8K9IpWzf0pl2umb13OG6B714ynLuma4tacQAvIxR1O2TX6i7G0VxLPviVRcg0KMhzWkIk7QNU5sRqBmaSB0vQjuHSis6FbSdGJVesaOwlAJ3dZg7LtKOAYMu28Wqc5KB1nLpNTM9FXCyd0IFhbf9bs8qfFKSbDaEEPLFJpD+60Jn8cuKgihX+B6tbg3BiYHhadZy3NbnMozUmr7f22CVhatA2ZxcTUsShYOEIuyPWRQtdi5ayqORz5HF1C16NO55rgboChvYMQxRiBlr4b8uWU6YWnyp++w51f4o+dGzqORnhTs+1uRDyVBslITwCNJ3qIZmEYbWCPIr+nCUbo62CjGfVDdoDnAzDkPxNdrd571cjm6AAdzAB+JUyOwEq0sjywUWfBNtGNyQuOZqAnhdiz+qxTVY+qXtOU7T6SOXHyQewK4FinnJc5xIIMUNl6A9GWn6pQ0J1TmpiMSKVF1cStwv0UsaGor6S\/qDm9MkSY57hvEMjOyOTsvmzewdGjnOZLis1g7VqDfgl9ENRC96UIdL1Qq4J9iJ6XyaH3Ns7nhc3qvybh5JJC5vxz0b3O8xK9TY8vEldL+8WRBuyG3tSRGaf\/BBNqlWnTiafI6cVo8jgQCZqCipjQOgzRH9TaO\/wSRcUXr+ARRm2kIwEMkFZVC9nUfnSAOcPWo4AZ0dzcP+CQiwAaDW5YKExBEMYMGfhv1fXsej95BciqNHqeKq3zhFJb8rQDgJj5C3T\/Nbq\/kFqOPvBa+Wg\/InrjVSNdOyRNN4wWXG54aa1pkR+fF6qcCYFxghqJSjqK2bqHqKZkRIuGJDzO++gr9GJo6tjbFLpCtDzWJoZjaAnBjvTjWJIdbWN3h07GBN9UAe6H31wgePJV2\/yb6dZxyn8+LhroyMYP2yxWLw==`,
	cmdQuit:                "",
}

// newLockListener creates a new listener on the local IP.
func newLocalListener() (net.Listener, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			return nil, fmt.Errorf("local listener: listen: %w", err)
		}
	}
	return l, nil
}

// server is a mock TeamSpeak 3 server.
type server struct {
	Addr     string
	Listener net.Listener

	wg        sync.WaitGroup
	noHeader  bool
	noBanner  bool
	failConn  bool
	badHeader bool
	useSSH    bool

	// Below here is protected by mtx.
	mtx    sync.Mutex
	conns  map[net.Conn]struct{}
	closed bool
	err    error
}

// sconn represents a server connection.
type sconn struct {
	net.Conn
}

type serverOption func(s *server)

func useSSH() serverOption {
	return func(s *server) {
		s.useSSH = true
	}
}

func noHeader() serverOption {
	return func(s *server) {
		s.noHeader = true
	}
}

func noBanner() serverOption {
	return func(s *server) {
		s.noBanner = true
	}
}

func failConn() serverOption {
	return func(s *server) {
		s.failConn = true
	}
}

func badHeader() serverOption {
	return func(s *server) {
		s.badHeader = true
	}
}

// newServer returns a running server. It fails the test immediately if an error occurred.
func newServer(t *testing.T, options ...serverOption) *server {
	t.Helper()
	l, err := newLocalListener()
	require.NoError(t, err)

	s := &server{
		Listener: l,
		conns:    make(map[net.Conn]struct{}),
	}
	for _, f := range options {
		f(s)
	}
	s.Addr = s.Listener.Addr().String()
	s.Start()

	return s
}

func (s *server) handleError(err error) bool {
	if err == nil {
		return false
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if !s.closed {
		s.err = err
	}

	return true
}

// Start starts the server.
func (s *server) Start() {
	s.wg.Add(1)
	go s.serve()
}

// singleClose ensures that a connection is only closed once
// to avoid spurious errors.
type singleClose struct {
	net.Conn

	once sync.Once
	err  error
}

func (c *singleClose) close() {
	c.err = c.Conn.Close()
}

func (c *singleClose) Close() error {
	c.once.Do(c.close)
	return c.err
}

// server processes incoming requests until signaled to stop with Close.
func (s *server) serve() {
	defer s.wg.Done()
	for {
		conn, err := s.Listener.Accept()
		if s.handleError(err) {
			return
		}

		if s.useSSH {
			conn, err = newSSHServerShell(&singleClose{Conn: conn})
			if s.handleError(err) {
				return
			}
		}
		s.wg.Add(1)
		go s.handle(conn)
	}
}

// writeResponse writes the given msg followed by an error (ok) response.
// If msg is empty the only the error (ok) rsponse is sent.
func (s *server) writeResponse(c *sconn, msg string) error {
	if msg != "" {
		if err := s.write(c.Conn, msg); err != nil {
			return err
		}
	}

	return s.write(c.Conn, errOK)
}

// write writes msg to w.
func (s *server) write(w io.Writer, msg string) error {
	if _, err := w.Write([]byte(msg + "\n\r")); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// handle handles a client connection.
func (s *server) handle(conn net.Conn) {
	s.mtx.Lock()
	s.conns[conn] = struct{}{}
	s.mtx.Unlock()

	defer func() {
		s.closeConn(conn)
		s.wg.Done()
	}()

	if s.failConn {
		return
	}

	sc := bufio.NewScanner(bufio.NewReader(conn))
	sc.Split(bufio.ScanLines)

	if !s.noHeader {
		if s.badHeader {
			if s.handleError(s.write(conn, "bad")) {
				return
			}
		} else {
			if s.handleError(s.write(conn, DefaultConnectHeader)) {
				return
			}
		}

		if !s.noBanner {
			if s.handleError(s.write(conn, banner)) {
				return
			}
		}
	}

	c := &sconn{Conn: conn}
	for sc.Scan() {
		l := sc.Text()
		parts := strings.Split(l, " ")
		cmd := strings.TrimSpace(parts[0])
		// Support server commands with specific optional parameters,
		// they can be bypassed from the usual parameter trimming here.
		if cmd == "clientlist" {
			cmd = l
		}
		resp, ok := commands[cmd]
		var err error
		switch {
		case ok:
			// Request has response, send it.
			err = s.writeResponse(c, resp)
		case cmd == "disconnect":
			return
		case cmd != "":
			err = s.write(c, errUnknownCmd)
		}

		if s.handleError(err) {
			return
		}

		if cmd == cmdQuit {
			return
		}
	}

	s.handleError(sc.Err())
}

// closeConn closes a client connection and removes it from our map of connections.
func (s *server) closeConn(conn net.Conn) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	conn.Close()
	delete(s.conns, conn)
}

// Close cleanly shuts down the server.
func (s *server) Close() error {
	s.mtx.Lock()
	s.closed = true
	err := s.Listener.Close()
	for c := range s.conns {
		if err2 := c.Close(); err2 != nil && err == nil {
			err = err2
		}
	}
	s.mtx.Unlock()
	s.wg.Wait()

	if err != nil {
		return err
	}

	return s.err
}

// sshServerShell provides an ssh server shell session.
type sshServerShell struct {
	net.Conn
	cond *sync.Cond

	// Everything below is protected by mtx.
	mtx        sync.RWMutex
	sshChannel ssh.Channel
	closed     bool
}

// newSSHServerShell creates a new sshServerShell from a net.Conn.
func newSSHServerShell(conn net.Conn) (*sshServerShell, error) {
	private, err := ssh.ParsePrivateKey([]byte(sshPrivateServerKey))
	if err != nil {
		return nil, fmt.Errorf("mock ssh shell: parse private key: %w", err)
	}

	config := &ssh.ServerConfig{NoClientAuth: true}
	config.AddHostKey(private)

	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return nil, fmt.Errorf("mock ssh shell: new server conn: %w", err)
	}
	go ssh.DiscardRequests(reqs)

	m := new(sync.Mutex)
	m.Lock()

	c := &sshServerShell{
		Conn: conn,
		cond: sync.NewCond(m),
	}

	go func() {
		newChan := <-chans
		if newChan.ChannelType() != "session" {
			_ = newChan.Reject(ssh.UnknownChannelType, ssh.UnknownChannelType.String())
		}

		sChan, reqs, _ := newChan.Accept()
		go func(in <-chan *ssh.Request) {
			for req := range in {
				_ = req.Reply(req.Type == "shell", nil)
			}
		}(reqs)

		c.mtx.Lock()
		c.sshChannel = sChan
		c.mtx.Unlock()
		c.cond.Broadcast()
	}()

	return c, nil
}

func (c *sshServerShell) channel() (ssh.Channel, bool) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	return c.sshChannel, c.closed
}

func (c *sshServerShell) waitChannel() (ssh.Channel, error) {
	for {
		ch, closed := c.channel()
		if closed {
			return nil, net.ErrClosed
		}
		if ch != nil {
			return c.sshChannel, nil
		}
		c.cond.Wait()
	}
}

// Read reads from the ssh channel.
func (c *sshServerShell) Read(b []byte) (int, error) {
	ch, err := c.waitChannel()
	if err != nil {
		return 0, err
	}

	return ch.Read(b) //nolint: wrapcheck
}

// Write writes to the ssh channel.
func (c *sshServerShell) Write(b []byte) (int, error) {
	ch, err := c.waitChannel()
	if err != nil {
		return 0, err
	}

	return ch.Write(b) //nolint: wrapcheck
}

// Close closes the ssh channel and connection.
func (c *sshServerShell) Close() error {
	c.mtx.Lock()
	c.closed = true
	c.mtx.Unlock()
	c.cond.Broadcast()

	return c.Conn.Close() //nolint: wrapcheck
}
