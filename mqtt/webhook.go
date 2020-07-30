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

type EventClientConnected struct {
	*EventBase
	ClientId    string `json:"clientid"`
	Username    string `json:"username"`
	KeepAlive   int    `json:"keepalive"`
	IpAddress   string `json:"ipaddress"`
	ProtoVer    int    `json:"proto_ver"`
	ConnectedAt int64  `json:"connected_at"`
	ConnAck     int    `json:"conn_ack"`
}

type EventClientDisconnected struct {
	*EventBase
	ClientId string `json:"clientid"`
	Username string `json:"username"`
	Reason   string `json:"reason"`
}

type EventClientSubscribe struct {
	*EventBase
	ClientId string `json:"clientid"`
	Username string `json:"username"`
	Topic    string `json:"topic"`
	Opts     struct {
		Qos int `json:"qos"`
	} `json:"opts"`
}

type EventClientUnsubscribe struct {
	*EventBase
	ClientId string `json:"clientid"`
	Username string `json:"username"`
	Topic    string `json:"topic"`
}

type EventSessionCreated struct {
	*EventBase
	ClientId string `json:"clientid"`
	Username string `json:"username"`
}

type EventSessionSubscribe struct {
	*EventBase
	ClientId string `json:"clientid"`
	Username string `json:"username"`
	Topic    string `json:"topic"`
	Opts     struct {
		Qos int `json:"qos"`
	} `json:"opts"`
}

type EventSessionUnsubscribe struct {
	*EventBase
	ClientId string `json:"clientid"`
	Username string `json:"username"`
	Topic    string `json:"topic"`
}

type EventSessionTerminated struct {
	*EventBase
	ClientId string `json:"clientid"`
	Username string `json:"username"`
	Reason   string `json:"reason"`
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
