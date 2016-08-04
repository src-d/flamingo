package flamingo

type OutgoingMessage struct {
	ChannelID   string
	Text        string
	Sender      *MessageSender
	Attachments []Attachment
}

func NewAttachment() Attachment {
	return &attachment{
		byClient: make(map[ClientType]interface{}),
	}
}

type Attachment interface {
	Clients() []ClientType
	ForClient(ClientType) (interface{}, bool)
	Add(ClientType, interface{}) Attachment
}

func (a *attachment) ForClient(client ClientType) (interface{}, bool) {
	a, ok := a.byClient[client]
	return a, ok
}

func (a *attachment) Add(client ClientType, attachment interface{}) Attachment {
	a.byClient[client] = attachment
	return a
}

func (a *attachment) Clients() []ClientType {
	clients := make([]ClientType, 0, len(a.byClient))
	for t := range a.byClient {
		clients = append(clients, t)
	}
	return clients
}

type Message struct {
	ID          string
	Type        ClientType
	User        User
	Channel     Channel
	Time        Time
	Attachments []Attachment
	Text        string
	Extra       interface{}
}

type Conversation []Message

type User struct {
	ID       string
	Username string
	Name     string
	IsBot    bool
	Type     ClientType
	Extra    interface{}
}

type Channel struct {
	ID    string
	Name  string
	IsDM  bool
	Users []User
	Type  ClientType
	Extra interface{}
}
