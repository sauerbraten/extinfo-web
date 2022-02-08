module github.com/sauerbraten/extinfo-web

go 1.13

require (
	github.com/gorilla/websocket v1.4.1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/sauerbraten/chef v0.0.0-20220203205700-ea2a4126f117
	github.com/sauerbraten/pubsub v0.0.0-20171021135711-897f02c5bb09
)

replace github.com/sauerbraten/chef => ../chef
