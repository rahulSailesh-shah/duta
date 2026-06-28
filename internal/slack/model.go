package slack

type Envelope struct {
	Token     string `json:"token"`
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Event     Event  `json:"event"`
	EventID   string `json:"event_id"`
	EventTime int64  `json:"event_time"`
}

type Event struct {
	Type        string `json:"type"`
	User        string `json:"user"`
	Text        string `json:"text"`
	Channel     string `json:"channel"`
	Ts          string `json:"ts"`
	ThreadTS    string `json:"thread_ts"`
	ClientMsgID string `json:"client_msg_id"`
	EventTs     string `json:"event_ts"`
}
