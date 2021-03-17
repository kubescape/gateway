module notification-server

go 1.13

replace asterix.cyberarmor.io/cyberarmor/capacketsgo => ./vendor/asterix.cyberarmor.io/cyberarmor/capacketsgo

require (
	asterix.cyberarmor.io/cyberarmor/capacketsgo v0.0.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/websocket v1.4.2
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
)
