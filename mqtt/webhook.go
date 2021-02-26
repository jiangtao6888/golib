package mqtt

const (
	EventTypeClientConnected     = "client_connected"
	EventTypeClientDisconnected  = "client_disconnected"
	EventTypeClientSubscribe     = "client_subscribe"
	EventTypeClientUnsubscribe   = "client_unsubscribe"
	EventTypeSessionCreated      = "session_created"
	EventTypeSessionSubscribe    = "session_subscribed"
	EventTypeSessionUnsubscribed = "session_unsubscribed"
	EventTypeSessionTerminated   = "session_terminated"
	EventTypeMessagePublish      = "message_publish"
	EventTypeMessageDeliver      = "message_deliver"
	EventTypeMessageAcked        = "message_acked"
	EventTypeMessageDropped      = "message_dropped"
)

type EventBase struct {
	Action string `json:"action"`
}

type EventMessageBase struct {
	Action  string `json:"action"`
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

type EventClientBase struct {
	*EventBase
	ClientId string `json:"clientid"`
	Username string `json:"username"`
}

type EventClientConnected struct {
	*EventClientBase
	KeepAlive   int    `json:"keepalive"`
	IpAddress   string `json:"ipaddress"`
	ProtoVer    int    `json:"proto_ver"`
	ConnectedAt int64  `json:"connected_at"`
	ConnAck     int    `json:"conn_ack"`
}

type EventClientDisconnected struct {
	*EventClientBase
	Reason string `json:"reason"`
}

type EventOpts struct {
	Qos int `json:"qos"`
}

type EventClientSubscribe struct {
	*EventClientBase
	Topic string     `json:"topic"`
	Opts  *EventOpts `json:"opts"`
}

type EventClientUnsubscribe struct {
	*EventClientBase
	Topic string `json:"topic"`
}

type EventSessionCreated EventClientBase

type EventSessionSubscribe struct {
	*EventClientBase
	Topic string     `json:"topic"`
	Opts  *EventOpts `json:"opts"`
}

type EventSessionUnsubscribe struct {
	*EventClientBase
	Topic string `json:"topic"`
}

type EventSessionTerminated struct {
	*EventClientBase
	Reason string `json:"reason"`
}

type EventMessagePublish struct {
	*EventMessageBase
	FromClientId string `json:"from_client_id"`
	FromUsername string `json:"from_username"`
	Qos          int    `json:"qos"`
	Retain       bool   `json:"retain"`
	Timestamp    int64  `json:"ts"`
}

type EventMessageDelivered struct {
	*EventMessageBase
	ClientId     string `json:"clientid"`
	Username     string `json:"username"`
	FromClientId string `json:"from_client_id"`
	FromUsername string `json:"from_username"`
	Qos          int    `json:"qos"`
	Retain       bool   `json:"retain"`
	Timestamp    int64  `json:"ts"`
}

type EventMessageAcked struct {
	*EventMessageBase
	ClientId     string `json:"clientid"`
	Username     string `json:"username"`
	FromClientId string `json:"from_client_id"`
	FromUsername string `json:"from_username"`
	Qos          int    `json:"qos"`
	Retain       bool   `json:"retain"`
	Timestamp    int64  `json:"ts"`
}

type EventMessageDropped struct {
	*EventMessageBase
	Username string `json:"username"`
}
